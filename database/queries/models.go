// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.14.0

package queries

import ()

type KvPair struct {
	K string
	V string
}

type Welcome struct {
	GuildID       string
	ChannelID     string
	MessageType   string
	MessageText   string
	ImageName     string
	ImageTitle    string
	ImageSubtitle string
}
