package domain

// Channel 通知渠道
type Channel string

const (
	ChannelSms   Channel = "sms"
	ChannelEmail Channel = "email"
)

func (c Channel) String() string {
	return string(c)
}

func (c Channel) IsValid() bool {
	switch c {
	case ChannelSms, ChannelEmail:
		return true
	}
	return false
}

func (c Channel) IsSms() bool {
	return c == ChannelSms
}

func (c Channel) IsEmail() bool {
	return c == ChannelEmail
}

// ActiveStatus 状态
type ActiveStatus string

const (
	ActiveStatusActive   ActiveStatus = "active"
	ActiveStatusInactive ActiveStatus = "inactive"
)

func (s ActiveStatus) String() string {
	return string(s)
}

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
