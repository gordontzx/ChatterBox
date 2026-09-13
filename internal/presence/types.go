package presence

import "time"

type Info struct {
	UserID	   string	  `json:"user_id"`
	Online	   bool		  `json:"online"`
	LastSeenAt *time.Time `json:"last_seen_at,omitempty"`
}

type Event struct {
	Type	   string	  `json:"type"`
	UserID	   string	  `json:"user_id"`
	Online	   bool		  `json:"online"`
	LastSeenAt *time.Time `json:"last_seen_at,omitempty"`
}
