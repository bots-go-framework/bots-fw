package invites

import "time"

const InviteKind = "Invite"

type Invite struct {
	Channel         string `datastore:",noindex"`
	For             string
	DtCreated       time.Time
	DtActivated     time.Time
	DtClaimed       time.Time
	CreatedByUserID int64
	ClaimedByUserID int64
	ClaimedUsing    string // What this? Need comment!

	ToName  string `datastore:",noindex"`
	ToEmail string
	ToSms   string
}
