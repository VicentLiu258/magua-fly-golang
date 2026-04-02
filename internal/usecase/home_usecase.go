package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"magua-fly-golang/internal/domain"
	"magua-fly-golang/internal/repository"
)

// ErrStudentNotFound 学员首页：当前用户无学员档案
var ErrStudentNotFound = errors.New("student not found")

// ErrInstructorNotFound 教员首页：当前用户无教员档案
var ErrInstructorNotFound = errors.New("instructor not found")

// ErrAdminUserNotFound 管理首页：用户不存在
var ErrAdminUserNotFound = errors.New("user not found")

func ptrRecordStatusNormal() *domain.TrainingRecordStatus {
	s := domain.RecordStatusNormal
	return &s
}

// StudentHomeRequest 学员首页可选参数（与设计稿 Query 对应，本仓库走 JSON body）
type StudentHomeRequest struct {
	Locale         string `json:"locale"`
	PostsLimit     int    `json:"posts_limit"`
	ArticlesLimit  int    `json:"articles_limit"`
}

// InstructorHomeRequest 教员首页可选参数
type InstructorHomeRequest struct {
	PostsLimit int `json:"posts_limit"`
}

// HomeUseCase 各角色首页聚合
type HomeUseCase struct {
	studentRepo   repository.StudentRepository
	userRepo      repository.UserRepository
	taskRepo      repository.TaskRepository
	archiveRepo   repository.TrainingArchiveRepository
	planRepo       repository.TrainingPlanRepository
	aircraftRepo   repository.AircraftRepository
	recordRepo     repository.TrainingRecordRepository
	instructorRepo repository.InstructorRepository
	trainingUC    *TrainingUseCase
	communityUC   *CommunityUseCase
	theoryUC      *TheoryUseCase
}

// NewHomeUseCase 创建首页聚合用例
func NewHomeUseCase(
	studentRepo repository.StudentRepository,
	userRepo repository.UserRepository,
	taskRepo repository.TaskRepository,
	archiveRepo repository.TrainingArchiveRepository,
	planRepo repository.TrainingPlanRepository,
	instructorRepo repository.InstructorRepository,
	aircraftRepo repository.AircraftRepository,
	recordRepo repository.TrainingRecordRepository,
	trainingUC *TrainingUseCase,
	communityUC *CommunityUseCase,
	theoryUC *TheoryUseCase,
) *HomeUseCase {
	return &HomeUseCase{
		studentRepo:    studentRepo,
		userRepo:       userRepo,
		taskRepo:       taskRepo,
		archiveRepo:    archiveRepo,
		planRepo:       planRepo,
		aircraftRepo:   aircraftRepo,
		recordRepo:     recordRepo,
		instructorRepo: instructorRepo,
		trainingUC:     trainingUC,
		communityUC:    communityUC,
		theoryUC:       theoryUC,
	}
}

func clampPostsLimit(n int) int {
	if n <= 0 {
		return 4
	}
	if n > 10 {
		return 10
	}
	return n
}

func clampArticlesLimit(n int) int {
	if n <= 0 {
		return 4
	}
	if n > 20 {
		return 20
	}
	return n
}

// StudentHome 学员首页 data 与 warnings（HTTP 仍由 handler 返回 200）
func (uc *HomeUseCase) StudentHome(ctx context.Context, userID string, req *StudentHomeRequest) (map[string]interface{}, []string, error) {
	if req == nil {
		req = &StudentHomeRequest{}
	}
	_ = req.Locale // 预留多语言

	stu, err := uc.studentRepo.GetByUserID(ctx, userID)
	if err != nil || stu == nil {
		return nil, nil, ErrStudentNotFound
	}

	var remaining *int
	if stu.ContractDurationMinutes != nil {
		r := *stu.ContractDurationMinutes - stu.FlownMinutes
		remaining = &r
	}
	profile := map[string]interface{}{
		"student_id":                stu.ID,
		"name":                      stu.Name,
		"student_no":                stu.StudentNo,
		"contract_duration_minutes": stu.ContractDurationMinutes,
		"flown_minutes":             stu.FlownMinutes,
		"remaining_minutes":         remaining,
		"training_progress":         stu.TrainingProgress,
	}

	status := domain.TaskStatusPending
	allTasks, _ := uc.taskRepo.ListByStudent(ctx, stu.ID, &status)
	pendingTasks := allTasks
	if len(allTasks) > 5 {
		pendingTasks = allTasks[:5]
	}
	tasksBlock := map[string]interface{}{
		"items": pendingTasks,
		"total": len(allTasks),
	}

	now := time.Now()
	from := now.Format("2006-01-02")
	to := now.AddDate(0, 0, 7).Format("2006-01-02")
	plans, _ := uc.planRepo.ListByDate(ctx, from, to)
	type upcomingItem struct {
		TrainingDate   string `json:"training_date"`
		StartTime      string `json:"start_time"`
		EndTime        string `json:"end_time"`
		InstructorName string `json:"instructor_name"`
		AircraftName   string `json:"aircraft_name"`
	}
	var upcoming []upcomingItem
	for _, p := range plans {
		if p.StudentID != stu.ID {
			continue
		}
		var insName, acName string
		if ins, err := uc.instructorRepo.GetByID(ctx, p.InstructorID); err == nil && ins != nil {
			insName = ins.Name
		}
		if ac, err := uc.aircraftRepo.GetByID(ctx, p.AircraftID); err == nil && ac != nil {
			acName = ac.Name
		}
		upcoming = append(upcoming, upcomingItem{
			TrainingDate:   p.PlanDate.Format("2006-01-02"),
			StartTime:      p.StartTime.Format("15:04"),
			EndTime:        p.EndTime.Format("15:04"),
			InstructorName: insName,
			AircraftName:   acName,
		})
	}

	archives, _ := uc.archiveRepo.ListByStudent(ctx, stu.ID, 1, 0)
	var lastArchive interface{}
	if len(archives) > 0 {
		a := archives[0]
		lastArchive = map[string]interface{}{
			"id":                      a.ID,
			"training_date":           a.TrainingDate.Format("2006-01-02"),
			"stage":                   a.Stage,
			"flight_duration_minutes": a.FlightDurationMinutes,
		}
	}

	theorySummary := map[string]interface{}{
		"completed_documents": 0,
		"avg_score":           0,
		"last_attempt_at":     nil,
	}
	subjectProgress := map[string]interface{}{
		"completed": 0,
		"total":     0,
	}

	var warnings []string

	// user 展示块（展示名优先学员姓名，头像来自账号）
	userBlock := map[string]interface{}{
		"id":           userID,
		"display_name": stu.Name,
		"avatar_url":   "",
	}
	if u, err := uc.userRepo.GetByID(ctx, userID); err == nil && u != nil {
		userBlock["id"] = u.ID
		if u.AvatarURL != "" {
			userBlock["avatar_url"] = u.AvatarURL
		}
		if strings.TrimSpace(stu.Name) == "" && u.Username != "" {
			userBlock["display_name"] = u.Username
		}
	}

	// 社区动态（广场）
	pl := clampPostsLimit(req.PostsLimit)
	var communityPosts []interface{}
	if uc.communityUC != nil {
		tab := string(domain.CommunityFeedGlobal)
		list, err := uc.communityUC.ListPosts(ctx, userID, &ListPostsRequest{
			Tab:      tab,
			PageSize: pl,
		})
		if err != nil {
			warnings = append(warnings, "community_posts_unavailable")
		} else {
			for _, it := range list.Items {
				if it.Post == nil {
					continue
				}
				p := it.Post
				authorName := ""
				var authorAvatar interface{}
				if it.Author != nil {
					authorName = it.Author.Username
					authorAvatar = it.Author.AvatarURL
					if authorAvatar == "" {
						authorAvatar = nil
					}
				}
				content := postContentPreview(p.Title, p.Body)
				communityPosts = append(communityPosts, map[string]interface{}{
					"id":            p.ID,
					"author_name":   authorName,
					"author_avatar": authorAvatar,
					"content":       content,
					"created_at":    p.CreatedAt.UTC().Format(time.RFC3339),
					"likes":         p.LikeCount,
					"comments":      p.CommentCount,
				})
			}
		}
	} else {
		warnings = append(warnings, "community_posts_unavailable")
	}

	// 推荐专题（理论文档前 N 条，按 sort_order）
	al := clampArticlesLimit(req.ArticlesLimit)
	var featuredArticles []interface{}
	if uc.theoryUC != nil {
		resp, err := uc.theoryUC.ListDocuments(ctx, userID, &ListDocumentsRequest{
			Page:     1,
			PageSize: al,
		})
		if err != nil {
			warnings = append(warnings, "featured_articles_unavailable")
		} else {
			for _, d := range resp.Items {
				if len(featuredArticles) >= al {
					break
				}
				featuredArticles = append(featuredArticles, map[string]interface{}{
					"id":              d.ID,
					"title":           d.Title,
					"category":        d.Category,
					"cover_image_url": "",
				})
			}
		}
	} else {
		warnings = append(warnings, "featured_articles_unavailable")
	}

	data := map[string]interface{}{
		"profile":               profile,
		"tasks":                 tasksBlock,
		"upcoming_training":     upcoming,
		"last_training_archive": lastArchive,
		"theory_summary":        theorySummary,
		"subject_progress":      subjectProgress,
		"user":                  userBlock,
		"weather_snippet":       nil,
		"community_posts":       communityPosts,
		"featured_articles":     featuredArticles,
		"featured_products":     []interface{}{},
		"notifications":         []interface{}{},
		"events":                []interface{}{},
		"warnings":              warnings,
	}
	return data, warnings, nil
}

func monthRange(t time.Time) (time.Time, time.Time) {
	loc := t.Location()
	start := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, loc)
	end := start.AddDate(0, 1, 0).Add(-time.Nanosecond)
	return start, end
}

func yearRange(t time.Time) (time.Time, time.Time) {
	loc := t.Location()
	start := time.Date(t.Year(), 1, 1, 0, 0, 0, 0, loc)
	end := time.Date(t.Year(), 12, 31, 23, 59, 59, 999999999, loc)
	return start, end
}

func sumFlightMinutes(records []*domain.TrainingRecord) int {
	var sum int
	for _, r := range records {
		if r.FlightDurationMinutes != nil {
			sum += *r.FlightDurationMinutes
		}
	}
	return sum
}

func listRecordsForInstructor(ctx context.Context, repo repository.TrainingRecordRepository, instructorID string, from, to time.Time) ([]*domain.TrainingRecord, error) {
	f := &domain.TrainingRecordListFilter{
		InstructorID: &instructorID,
		Status:       ptrRecordStatusNormal(),
		DateFrom:     &from,
		DateTo:       &to,
	}
	items, _, err := repo.List(ctx, f, nil, 10000, 0)
	return items, err
}

// InstructorHome 教员首页
func (uc *HomeUseCase) InstructorHome(ctx context.Context, userID string, req *InstructorHomeRequest) (map[string]interface{}, []string, error) {
	if req == nil {
		req = &InstructorHomeRequest{}
	}
	ins, err := uc.instructorRepo.GetByUserID(ctx, userID)
	if err != nil || ins == nil {
		return nil, nil, ErrInstructorNotFound
	}
	u, _ := uc.userRepo.GetByID(ctx, userID)
	displayName := ins.Name
	if u != nil && strings.TrimSpace(displayName) == "" {
		displayName = u.Username
	}
	avatar := ""
	if u != nil {
		avatar = u.AvatarURL
	}

	var warnings []string
	now := time.Now()
	mStart, mEnd := monthRange(now)
	yStart, yEnd := yearRange(now)

	monthlyRec, err := listRecordsForInstructor(ctx, uc.recordRepo, ins.ID, mStart, mEnd)
	if err != nil {
		warnings = append(warnings, "instructor_stats_partial")
		monthlyRec = nil
	}
	annualRec, err := listRecordsForInstructor(ctx, uc.recordRepo, ins.ID, yStart, yEnd)
	if err != nil {
		warnings = append(warnings, "instructor_stats_partial")
		annualRec = nil
	}

	monthlyMin := sumFlightMinutes(monthlyRec)
	annualMin := sumFlightMinutes(annualRec)
	monthlyHours := float64(monthlyMin) / 60.0
	annualHours := float64(annualMin) / 60.0

	// 在训人数：近 365 天与该教员有训练记录的去重学员数
	yAgo := now.AddDate(-1, 0, 0)
	allYearRec, _ := listRecordsForInstructor(ctx, uc.recordRepo, ins.ID, yAgo, now)
	stuSet := map[string]struct{}{}
	for _, r := range allYearRec {
		if r.StudentID != "" {
			stuSet[r.StudentID] = struct{}{}
		}
	}
	activeStudents := len(stuSet)

	dist := uc.flightTimeDistribution(annualRec)
	if len(dist) == 0 {
		dist = []map[string]interface{}{
			{"type": "dual", "label": "带飞 (Dual)", "percent": 0},
			{"type": "xc", "label": "转场 (XC)", "percent": 0},
			{"type": "ifr", "label": "仪表 (IFR)", "percent": 0},
			{"type": "night", "label": "夜航 (Night)", "percent": 0},
		}
	}

	pl := clampPostsLimit(req.PostsLimit)
	var myPosts []interface{}
	if uc.communityUC != nil {
		list, err := uc.communityUC.ListPosts(ctx, userID, &ListPostsRequest{
			AuthorUserID: userID,
			PageSize:     pl,
		})
		if err != nil {
			warnings = append(warnings, "my_posts_unavailable")
		} else {
			for _, it := range list.Items {
				if it.Post == nil {
					continue
				}
				p := it.Post
				myPosts = append(myPosts, map[string]interface{}{
					"id":         p.ID,
					"content":    postContentPreview(p.Title, p.Body),
					"image_url":  firstMediaURL(p.MediaURLs),
					"created_at": p.CreatedAt.UTC().Format(time.RFC3339),
					"likes":      p.LikeCount,
					"comments":   p.CommentCount,
				})
			}
		}
	} else {
		warnings = append(warnings, "my_posts_unavailable")
	}

	data := map[string]interface{}{
		"profile": map[string]interface{}{
			"user_id":      userID,
			"display_name": displayName,
			"avatar_url":   avatar,
			"title":        "飞行教员",
			"tagline":      fmt.Sprintf("本月飞行 %.1f 小时", monthlyHours),
			"rank": map[string]interface{}{
				"code":               "FI",
				"label":              "飞行教员",
				"progress_percent":   0,
				"hint":               "等级进度待对接教员档案表",
			},
		},
		"stats": map[string]interface{}{
			"annual_flight_hours":     round1(annualHours),
			"monthly_flight_hours":    round1(monthlyHours),
			"active_students_count":   activeStudents,
			"passing_rate_percent":    0,
			"trend_annual":            "",
			"trend_monthly":             "",
			"trend_students":            "",
			"trend_passing":             "",
		},
		"flight_time_distribution": dist,
		"my_posts":                 myPosts,
		"warnings":                 warnings,
	}
	return data, warnings, nil
}

func postContentPreview(title, body string) string {
	title = strings.TrimSpace(title)
	body = strings.TrimSpace(body)
	if title == "" {
		return body
	}
	if body == "" {
		return title
	}
	return title + "\n" + body
}

func firstMediaURL(urls []string) interface{} {
	for _, s := range urls {
		if strings.TrimSpace(s) != "" {
			return s
		}
	}
	return nil
}

func round1(f float64) float64 {
	return float64(int(f*10+0.5)) / 10
}

func (uc *HomeUseCase) flightTimeDistribution(records []*domain.TrainingRecord) []map[string]interface{} {
	var dual, xc, ifr, night int
	for _, r := range records {
		m := 0
		if r.FlightDurationMinutes != nil {
			m = *r.FlightDurationMinutes
		}
		if m <= 0 {
			continue
		}
		ft := strings.ToLower(r.FlightType + " " + r.TaskType + " " + r.Destination + " " + r.Departure)
		n := strings.Contains(ft, "夜") || strings.Contains(ft, "night")
		i := strings.Contains(ft, "仪表") || strings.Contains(ft, "ifr")
		x := strings.Contains(ft, "转场") || strings.Contains(ft, "xc") || (!r.IsLocal && (r.Destination != "" || r.Departure != ""))
		switch {
		case n:
			night += m
		case i:
			ifr += m
		case x:
			xc += m
		default:
			dual += m
		}
	}
	total := dual + xc + ifr + night
	if total == 0 {
		return nil
	}
	pct := func(v int) int {
		return int((float64(v)*100.0)/float64(total) + 0.5)
	}
	return []map[string]interface{}{
		{"type": "dual", "label": "带飞 (Dual)", "percent": pct(dual)},
		{"type": "xc", "label": "转场 (XC)", "percent": pct(xc)},
		{"type": "ifr", "label": "仪表 (IFR)", "percent": pct(ifr)},
		{"type": "night", "label": "夜航 (Night)", "percent": pct(night)},
	}
}

// AdminHome 管理后台首页（KPI 来自训练记录聚合；举报/合规为占位）
func (uc *HomeUseCase) AdminHome(ctx context.Context, adminUserID string) (map[string]interface{}, []string, error) {
	u, err := uc.userRepo.GetByID(ctx, adminUserID)
	if err != nil || u == nil {
		return nil, nil, ErrAdminUserNotFound
	}
	displayName := u.Username

	now := time.Now()
	mStart, mEnd := monthRange(now)
	filter := &domain.TrainingRecordListFilter{
		Status:   ptrRecordStatusNormal(),
		DateFrom: &mStart,
		DateTo:   &mEnd,
	}
	items, _, err := uc.recordRepo.List(ctx, filter, nil, 50000, 0)
	var warnings []string
	if err != nil {
		warnings = append(warnings, "kpi_partial")
		items = nil
	}
	monthlyMin := sumFlightMinutes(items)
	monthlyHours := float64(monthlyMin) / 60.0

	// 在训学员数：全表学员数量（简单实现）
	activeStudents := 0
	const pageSize = 500
	for off := 0; ; off += pageSize {
		batch, err := uc.studentRepo.List(ctx, pageSize, off, false)
		if err != nil {
			warnings = append(warnings, "active_students_partial")
			break
		}
		activeStudents += len(batch)
		if len(batch) < pageSize {
			break
		}
	}

	data := map[string]interface{}{
		"admin_user": map[string]interface{}{
			"id":           adminUserID,
			"display_name": displayName,
		},
		"kpi": map[string]interface{}{
			"monthly_total_flight_hours":       int(monthlyHours + 0.5),
			"monthly_total_flight_hours_trend": "",
			"active_students_count":            activeStudents,
			"active_students_delta":            "",
			"pending_reports_count":            0,
		},
		"modules": defaultAdminModules(),
		"system_alerts": map[string]interface{}{
			"summary": "",
			"items":   []interface{}{},
		},
		"warnings": warnings,
	}
	return data, warnings, nil
}

func defaultAdminModules() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"key":         "student_management",
			"title":       "学员管理",
			"description": "学员档案、进度、合同与权益",
			"path":        "/admin/student-management",
			"icon":        "graduation_cap",
		},
	}
}
