ALTER TYPE user_role ADD VALUE 'maintenance';
COMMENT ON TYPE user_role IS '用户角色枚举：admin（管理员）、staff（工作人员）、maintenance（机务）、instructor（教员）、student（学员）';
