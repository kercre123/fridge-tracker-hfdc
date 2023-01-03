# Hunger Free Dallas County fridge tracker software

This repo contains the software used to track community fridge usage by Hunger Free Dallas County (The Food Grid).

## Overview

-	"Fridge Grids" (ESP8266 trackers) make a "status" request to the server when they first turn on
	-	They get added to a fridgeList.json
	-	They send status updates every ~5 minutes
	-	Every status update is logged as CSV to ./webroot/statusLogs/`name`.csv
-	When their distance sensors detect an open fridge door, they will begin a timer. When the door is closed, the timer stops and an "open" request is made
	-	Every open requests is logged as CSV to ./webroot/openLogs/`name`.csv
-	Every request also includes a temperature and humidity reading of the inside of the fridge
-	It hosts a basic data-viewing site (at /)
	-	A basic dropdown is given and you can view:
		-	How many times a certain fridge has been opened
		-	How long it has been in an opened state
		-	A (basic) chart
		-	Links to download the raw CSV files
-	There is preliminary support for sending emails to owners of community fridges whenever contact is lost
	-	A timer is made for each fridge
		-	Timers are saved to JSON every 5 minutes in case of program shutdown
	-	A fridge is considered down if no status update has been sent in more than 30 minutes

## Why

Community fridges are obtained partially through the help of grants from the county. Having usage data for the fridges is very useful in applying for these grants.

## Technical details

-	Server software written with Go 1.18
-	Tracker software written with Arduino IDE
-	Tracker (client) hardware used:
	-	WeMos D1 mini ESP8266
	-	ST VL53L0X distance sensor
	-	DHT22 temperature sensor
-	Arduino IDE libraries used:
	-	Adafruit Unified Sensor
	-	Adafruit_VL53L0X
	-	DHT sensor library (by Adafruit)
-	Server software env vars:
	-	MAIL_EMAIL
		-	"From" email address
	-	MAIL_PASSWORD
		-	If using gmail: get an App Password from security settings
	-	ADMIN_EMAIL
		-	Secondary "To" email address
		-	Emergency email is sent to admin as well as fridge owner
	-	PORT
		-	Default is 83
