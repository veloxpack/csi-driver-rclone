# calendar

Go calendar Date (only), clock Time (only), and nullable versions of same, with sensible JSON and SQL behavior.

I wrote this because I needed:

 * calendar dates (with no associated time) 
 * clock times (with no associated date)
 * with no time zone (time zone to be stored separately as an Olsen TZDB value)
 * with support for nullable dates/times
 * compatible with PostgreSQL DATE and TIME types, and
 * with sensible ISO-8601 bare-date and bare-time JSON representations
 
So I can define a structure with (say):

```
StartDate         calendar.Date     `json:"start_date"`
StartTime         calendar.NullTime `json:"start_time"`
EndDate           calendar.Date     `json:"end_date"`
EndTime           calendar.NullTime `json:"end_time"`
TimeZone          string            `json:"timezone"`	
```

and scan in values from a PostgreSQL database, and serialize them to JSON, and get out something like:

…`"start_date":"2019-02-14","start_time":"09:00:00","end_date":"2019-02-14","end_time":"17:00:00","timezone":"Americas/Chicago"`…

If you don't have those precise requirements, this is not the date/time library for you.

