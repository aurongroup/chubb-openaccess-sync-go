package models

import "errors"

var (
	// AccessRecord
	ErrAccessRecordMissingLast      = errors.New("access record: missing required Last")
	ErrAccessRecordMissingBadgeID   = errors.New("access record: missing required BadgeID")
	ErrAccessRecordMissingStatus    = errors.New("access record: missing required Status")
	ErrAccessRecordMissingBadgeType = errors.New("access record: missing required BadgeType")

	// LnlBadgeStatus
	ErrBadgeStatusMissingID   = errors.New("badge status: missing required ID")
	ErrBadgeStatusMissingName = errors.New("badge status: missing required Name")

	// LnlBadgeType
	ErrBadgeTypeMissingID   = errors.New("badge type: missing required ID")
	ErrBadgeTypeMissingName = errors.New("badge type: missing required Name")

	// LnlAccessLevel
	ErrAccessLevelMissingID   = errors.New("access level: missing required ID")
	ErrAccessLevelMissingName = errors.New("access level: missing required Name")

	// LnlCardholder
	ErrCardholderMissingIdentifier = errors.New("cardholder: must have an ID or SSNO")
	ErrCardholderMissingLastName   = errors.New("cardholder: missing required LastName")

	// LnlBadge
	ErrBadgeNilCache         = errors.New("badge: cache is nil")
	ErrBadgeMissingID        = errors.New("badge: missing required ID")
	ErrBadgeMissingBadgeKey  = errors.New("badge: missing required BadgeKey")
	ErrBadgeUnresolvedStatus = errors.New("badge: STATUS not found in cache")
	ErrBadgeUnresolvedType   = errors.New("badge: TYPE not found in cache")

	// LnlAccessLevelAssignment
	ErrAssignmentNilCache              = errors.New("assignment: cache is nil")
	ErrAssignmentUnresolvedAccessLevel = errors.New("assignment: AccessLevelID not found in cache")
	ErrAssignmentUnresolvedBadge       = errors.New("assignment: BadgeKey not found in cache")
)
