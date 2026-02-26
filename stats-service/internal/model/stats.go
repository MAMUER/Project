package model

import "time"

type VisitStats struct {
    TotalVisits   int       `json:"totalVisits"`
    UniqueMembers int       `json:"uniqueMembers"`
    Date          time.Time `json:"date"`
}

type MemberActivity struct {
    MemberID   int    `json:"memberId"`
    MemberName string `json:"memberName"`
    VisitCount int    `json:"visitCount"`
    LastVisit  string `json:"lastVisit"`
}

type ClubStats struct {
    ClubID      int    `json:"clubId"`
    ClubName    string `json:"clubName"`
    TodayVisits int    `json:"todayVisits"`
    WeekVisits  int    `json:"weekVisits"`
    MonthVisits int    `json:"monthVisits"`
}