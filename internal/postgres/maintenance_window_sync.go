package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"magua-fly-golang/internal/domain"
	"magua-fly-golang/internal/repository"
)

var _ repository.MaintenanceWindowSyncer = (*DB)(nil)

const maintenanceRejectReason = "飞机维护/定检，时段已取消"

func maxTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

func minTime(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}

// dailyMaintenanceSegments 将 [startAt,endAt] 按自然日拆成多条 availability 记录（DATE + TIME）
func dailyMaintenanceSegments(startAt, endAt time.Time) []struct {
	dayDate time.Time
	start   time.Time
	end     time.Time
} {
	if !endAt.After(startAt) {
		return nil
	}
	loc := startAt.Location()
	var out []struct {
		dayDate time.Time
		start   time.Time
		end     time.Time
	}
	d0 := time.Date(startAt.Year(), startAt.Month(), startAt.Day(), 0, 0, 0, 0, loc)
	d1 := time.Date(endAt.Year(), endAt.Month(), endAt.Day(), 0, 0, 0, 0, loc)
	for d := d0; !d.After(d1); d = d.AddDate(0, 0, 1) {
		dayStart := d
		dayEnd := d.Add(24*time.Hour).Add(-time.Nanosecond)
		segS := maxTime(startAt, dayStart)
		segE := minTime(endAt, dayEnd)
		if segE.After(segS) {
			out = append(out, struct {
				dayDate time.Time
				start   time.Time
				end     time.Time
			}{dayDate: dayStart, start: segS, end: segE})
		}
	}
	return out
}

func clockForDB(t time.Time) time.Time {
	return time.Date(0, 1, 1, t.Hour(), t.Minute(), t.Second(), 0, time.UTC)
}

// bookingOverlapsMaintenance 判断预约与维护窗在绝对时间上是否重叠
func bookingOverlapsMaintenance(b *domain.Booking, winStart, winEnd time.Time) bool {
	loc := winStart.Location()
	pd := b.PlanDate
	bs := time.Date(pd.Year(), pd.Month(), pd.Day(), b.StartTime.Hour(), b.StartTime.Minute(), b.StartTime.Second(), 0, loc)
	be := time.Date(pd.Year(), pd.Month(), pd.Day(), b.EndTime.Hour(), b.EndTime.Minute(), b.EndTime.Second(), 0, loc)
	if !be.After(bs) {
		be = be.AddDate(0, 0, 1)
	}
	return bs.Before(winEnd) && winStart.Before(be)
}

// CreateMaintenanceWindowWithSync 创建维护窗、写入 availability，可选同步拒绝冲突预约并删除对应 training_plans。
// 检修排期仅由人工维护窗驱动；机时定检字段不参与此流程。
func (db *DB) CreateMaintenanceWindowWithSync(ctx context.Context, aircraftID, kind string, startAt, endAt time.Time, hangar, notes string, syncSchedule bool, reviewerID string) (*domain.MaintenanceWindowSyncResult, error) {
	if kind == "" {
		kind = domain.DefaultMaintenanceWindowKind
	}
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var wid string
	var cAt, uAt time.Time
	err = tx.QueryRow(ctx, `
		INSERT INTO aircraft_maintenance_windows (
			aircraft_id, kind, scheduled_start_at, scheduled_end_at, status, hangar_location, notes
		) VALUES ($1::uuid, $2, $3, $4, 'planned', NULLIF($5,''), NULLIF($6,''))
		RETURNING id::text, created_at, updated_at`,
		aircraftID, kind, startAt, endAt, hangar, notes,
	).Scan(&wid, &cAt, &uAt)
	if err != nil {
		return nil, err
	}

	segs := dailyMaintenanceSegments(startAt, endAt)
	for _, seg := range segs {
		_, err = tx.Exec(ctx, `
			INSERT INTO aircraft_availability (aircraft_id, date, start_time, end_time, status, maintenance_window_id)
			VALUES ($1::uuid, $2::date, $3, $4, 'maintenance', $5::uuid)`,
			aircraftID, seg.dayDate, clockForDB(seg.start), clockForDB(seg.end), wid,
		)
		if err != nil {
			return nil, err
		}
	}

	result := &domain.MaintenanceWindowSyncResult{
		Window: &domain.MaintenanceWindow{
			ID:               wid,
			AircraftID:       aircraftID,
			Kind:             kind,
			ScheduledStartAt: startAt,
			ScheduledEndAt:   endAt,
			Status:           domain.MaintenancePlanned,
			HangarLocation:   hangar,
			Notes:            notes,
			CreatedAt:        cAt,
			UpdatedAt:        uAt,
		},
	}

	if syncSchedule {
		affected, plans, err := syncBookingsForMaintenance(ctx, tx, aircraftID, startAt, endAt, reviewerID)
		if err != nil {
			return nil, err
		}
		result.AffectedBookings = affected
		result.CancelledPlans = plans
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return result, nil
}

func syncBookingsForMaintenance(ctx context.Context, tx pgx.Tx, aircraftID string, winStart, winEnd time.Time, reviewerID string) (affected []string, planIDs []string, err error) {
	fromD := time.Date(winStart.Year(), winStart.Month(), winStart.Day(), 0, 0, 0, 0, winStart.Location())
	toD := time.Date(winEnd.Year(), winEnd.Month(), winEnd.Day(), 0, 0, 0, 0, winEnd.Location())
	rows, err := tx.Query(ctx, `
		SELECT id::text, student_id::text, instructor_id::text, aircraft_id::text, plan_date, start_time, end_time, status
		FROM bookings
		WHERE aircraft_id = $1::uuid
		  AND plan_date >= $2::date AND plan_date <= $3::date
		  AND status IN ('pending', 'approved')
		ORDER BY plan_date, start_time`,
		aircraftID, fromD, toD,
	)
	if err != nil {
		return nil, nil, err
	}
	var candidates []domain.Booking
	for rows.Next() {
		var b domain.Booking
		if err := rows.Scan(&b.ID, &b.StudentID, &b.InstructorID, &b.AircraftID, &b.PlanDate, &b.StartTime, &b.EndTime, &b.Status); err != nil {
			rows.Close()
			return nil, nil, err
		}
		candidates = append(candidates, b)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, nil, err
	}
	rows.Close()

	var reviewer interface{}
	if reviewerID != "" {
		reviewer = reviewerID
	}

	for i := range candidates {
		b := &candidates[i]
		if !bookingOverlapsMaintenance(b, winStart, winEnd) {
			continue
		}
		if b.Status == domain.BookingApproved {
			var pid string
			qerr := tx.QueryRow(ctx, `
				DELETE FROM training_plans
				WHERE student_id = $1::uuid AND instructor_id = $2::uuid AND aircraft_id = $3::uuid
				  AND plan_date = $4::date AND start_time = $5 AND end_time = $6
				RETURNING id::text`,
				b.StudentID, b.InstructorID, b.AircraftID, b.PlanDate, b.StartTime, b.EndTime,
			).Scan(&pid)
			if qerr != nil {
				if !errors.Is(qerr, pgx.ErrNoRows) {
					return nil, nil, qerr
				}
			} else {
				planIDs = append(planIDs, pid)
			}
		}
		_, execErr := tx.Exec(ctx, `
			UPDATE bookings SET status = 'rejected', reject_reason = $2, reviewed_at = NOW(),
				reviewer_id = $3, updated_at = NOW()
			WHERE id = $1::uuid`,
			b.ID, maintenanceRejectReason, reviewer,
		)
		if execErr != nil {
			return nil, nil, execErr
		}
		affected = append(affected, b.ID)
	}
	return affected, planIDs, nil
}

// UpdateMaintenanceWindowWithSync 更新维护窗时间/备注并替换 availability，重新执行排班同步
func (db *DB) UpdateMaintenanceWindowWithSync(ctx context.Context, windowID, aircraftID, kind string, startAt, endAt time.Time, hangar, notes string, syncSchedule bool, reviewerID string) (*domain.MaintenanceWindowSyncResult, error) {
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var oldAircraft string
	err = tx.QueryRow(ctx, `SELECT aircraft_id::text FROM aircraft_maintenance_windows WHERE id = $1::uuid`, windowID).Scan(&oldAircraft)
	if err != nil {
		return nil, err
	}
	if aircraftID != "" && aircraftID != oldAircraft {
		return nil, fmt.Errorf("aircraft_id mismatch")
	}
	aircraftID = oldAircraft

	_, err = tx.Exec(ctx, `DELETE FROM aircraft_availability WHERE maintenance_window_id = $1::uuid`, windowID)
	if err != nil {
		return nil, err
	}

	k := kind
	if k == "" {
		k = domain.DefaultMaintenanceWindowKind
	}
	tag, err := tx.Exec(ctx, `
		UPDATE aircraft_maintenance_windows SET
			kind = $2,
			scheduled_start_at = $3,
			scheduled_end_at = $4,
			hangar_location = NULLIF($5,''),
			notes = NULLIF($6,''),
			updated_at = NOW()
		WHERE id = $1::uuid AND status <> 'cancelled'::maintenance_window_status`,
		windowID, k, startAt, endAt, hangar, notes,
	)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, fmt.Errorf("maintenance window not found or already cancelled")
	}

	segs := dailyMaintenanceSegments(startAt, endAt)
	for _, seg := range segs {
		_, err = tx.Exec(ctx, `
			INSERT INTO aircraft_availability (aircraft_id, date, start_time, end_time, status, maintenance_window_id)
			VALUES ($1::uuid, $2::date, $3, $4, 'maintenance', $5::uuid)`,
			aircraftID, seg.dayDate, clockForDB(seg.start), clockForDB(seg.end), windowID,
		)
		if err != nil {
			return nil, err
		}
	}

	var w domain.MaintenanceWindow
	var hangarPtr, notesPtr *string
	err = tx.QueryRow(ctx, `
		SELECT id::text, aircraft_id::text, kind, scheduled_start_at, scheduled_end_at, status::text,
		       hangar_location, notes, created_at, updated_at
		FROM aircraft_maintenance_windows WHERE id = $1::uuid`, windowID,
	).Scan(&w.ID, &w.AircraftID, &w.Kind, &w.ScheduledStartAt, &w.ScheduledEndAt, &w.Status, &hangarPtr, &notesPtr, &w.CreatedAt, &w.UpdatedAt)
	if err != nil {
		return nil, err
	}
	if hangarPtr != nil {
		w.HangarLocation = *hangarPtr
	}
	if notesPtr != nil {
		w.Notes = *notesPtr
	}

	result := &domain.MaintenanceWindowSyncResult{Window: &w}

	if syncSchedule {
		affected, plans, err := syncBookingsForMaintenance(ctx, tx, aircraftID, startAt, endAt, reviewerID)
		if err != nil {
			return nil, err
		}
		result.AffectedBookings = affected
		result.CancelledPlans = plans
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return result, nil
}

// CancelMaintenanceWindow 取消维护计划并删除关联 availability
func (db *DB) CancelMaintenanceWindow(ctx context.Context, windowID string) error {
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `DELETE FROM aircraft_availability WHERE maintenance_window_id = $1::uuid`, windowID)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `
		UPDATE aircraft_maintenance_windows SET status = 'cancelled'::maintenance_window_status, updated_at = NOW()
		WHERE id = $1::uuid`, windowID)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}
