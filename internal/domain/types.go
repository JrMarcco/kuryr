package domain

// ActiveStatus 状态
type ActiveStatus string

// AuditStatus 审核状态
type AuditStatus string

const (
	AuditStatusPending  AuditStatus = "pending"  // 待审核
	AuditStatusAuditing AuditStatus = "auditing" // 审核中
	AuditStatusApproved AuditStatus = "approved" // 审核通过
	AuditStatusRejected AuditStatus = "rejected" // 审核拒绝
)

func (s AuditStatus) String() string {
	return string(s)
}

func (s AuditStatus) IsValid() bool {
	switch s {
	case AuditStatusPending, AuditStatusAuditing, AuditStatusApproved, AuditStatusRejected:
		return true
	}
	return false
}

func (s AuditStatus) IsPending() bool {
	return s == AuditStatusPending
}

func (s AuditStatus) IsAuditing() bool {
	return s == AuditStatusAuditing
}

func (s AuditStatus) IsApproved() bool {
	return s == AuditStatusApproved
}

func (s AuditStatus) IsRejected() bool {
	return s == AuditStatusRejected
}
