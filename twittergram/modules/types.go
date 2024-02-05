package modules

type TwitterAPIData struct {
	Data *struct {
		User *struct {
			Result struct {
				Legacy Legacy `json:"legacy"`
			} `json:"result"`
		} `json:"user,omitempty"`
	} `json:"data"`
}
type Legacy struct {
	FullText         string           `json:"full_text"`
	ExtendedEntities ExtendedEntities `json:"extended_entities"`
	Name             string           `json:"name"`
	Description      string           `json:"description"`
	Verified       bool   `json:"verified"`
	FollowersCount int    `json:"followers_count"`
	FriendsCount   int    `json:"friends_count"`
	StatusesCount  int    `json:"statuses_count"`
	Location       string `json:"location"`
}
