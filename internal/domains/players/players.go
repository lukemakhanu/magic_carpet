package players

import (
	"fmt"
	"time"
)

/* CREATE TABLE `players` (
  `prayer_id` bigint(30) NOT NULL,
  `profile_tag` varchar(400) NOT NULL,
  `status` enum('active','inactive','suspended','blocked') NOT NULL,
  `created` datetime NOT NULL,
  `modified` datetime NOT NULL DEFAULT current_timestamp()
);
*/

// NewPlayers instantiate players Struct
func NewPlayers(profileTag, status string) (*Players, error) {

	if profileTag == "" {
		return &Players{}, fmt.Errorf("profileTag not set")
	}

	if status == "" {
		return &Players{}, fmt.Errorf("status not set")
	}

	created := time.Now().Format("2006-01-02 15:04:05")
	modified := time.Now().Format("2006-01-02 15:04:05")

	return &Players{
		ProfileTag: profileTag,
		Status:     status,
		Created:    created,
		Modified:   modified,
	}, nil
}
