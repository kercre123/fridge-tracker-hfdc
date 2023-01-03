#include <ESP8266WiFi.h>
#include "Adafruit_VL53L0X.h"
#include <ESP8266HTTPClient.h>
#include <ESP8266WebServer.h>
#include <EEPROM.h>
#include "DHT.h"

#define DHTPIN 2
#define DHTTYPE DHT22

// much of the code handling wifi credentials is from how2electronics.com

// the ESP sets up an AP for entering wifi credentials
String apName = "FridgeHFDCSetup";
String apPassword = "aaaaaaaa";

// change these URLs to point to your server
String openURL = "http://SERVER_URL:83/api/open";
String statusURL = "http://SERVER_URL:83/api/status";

// will get filled in by EEPROM
String fridgeName = "";

// Time between each status update
float statusWaitTime = 300;

// some vars for eeprom reading and http handling
String st;
String content;
int i = 0;
int statusCode;

Adafruit_VL53L0X lox = Adafruit_VL53L0X();
WiFiClient wifi;
float timerFloat = 0;
int timer = 0;
int isOpen = 0;
int notSent = 0;
int bootUp = 1;
float statusTimer = 0;
String isOpenString = "no";
DHT dht(DHTPIN, DHTTYPE);
 
//Function Decalration
bool testWifi(void);
void launchWeb(void);
void setupAP(void);
 
//Establishing Local server at port 80 whenever required
ESP8266WebServer server(80);
 
void setup()
{
  Serial.begin(115200);
  Serial.println("Delaying for 5 seconds...");
  for (int i = 0; i < 6; ++i)
  {
    delay(1000);
    Serial.println(i);
  }
  Serial.println("Initializing ESP");
  WiFi.disconnect();
  EEPROM.begin(512);
  delay(10);
  pinMode(LED_BUILTIN, OUTPUT);
  Serial.println();
  Serial.println();
  Serial.println("Reading EEPROM for SSID, password, and fridge name");
 
  String esid;
  for (int i = 0; i < 32; ++i)
  {
    esid += char(EEPROM.read(i));
  }
  Serial.println();
  Serial.print("Saved SSID: `");
  Serial.print(esid.c_str());
  Serial.println("`");
 
  String epass = "";
  for (int i = 32; i < 64; ++i)
  {
    epass += char(EEPROM.read(i));
  }
  Serial.print("Saved password: `");
  Serial.print(epass.c_str());
  Serial.println("`");
 
  String ename = "";
  for (int i = 64; i < 96; ++i)
  {
    ename += char(EEPROM.read(i));
  }
  fridgeName = ename.c_str();
  Serial.print("Saved fridge name: `");
  Serial.print(fridgeName);
  Serial.println("`");
  
 
 
  WiFi.begin(esid.c_str(), epass.c_str());
  if (testWifi())
  {
    Serial.println("Wi-Fi connected, initializing sensors and tracking");
  // put code
    lox.begin();
    dht.begin();
    return;
  }
  else
  {
    Serial.println("Wi-Fi not connected, starting webserver and AP");
    launchWeb();
    setupAP();// Setup HotSpot
  }
 
  Serial.println();
  Serial.println("Waiting.");
  
  while ((WiFi.status() != WL_CONNECTED))
  {
    Serial.print(".");
    delay(100);
    server.handleClient();
  }
 
}
void loop()
{
  VL53L0X_RangingMeasurementData_t measure;
  lox.rangingTest(&measure, false);
  if (measure.RangeMilliMeter > 70) {
    // makeshift timer, this loop runs every ~.1 seconds
      isOpen = 1;
      notSent = 1;
      timerFloat = timerFloat + 0.13;
  } else {
    if (timerFloat > 1.50) {
    if (notSent == 1) {
       if (WiFi.status() == WL_CONNECTED) {
    HTTPClient http;
    http.begin(wifi, openURL);
    http.addHeader("Content-Type", "application/x-www-form-urlencoded");
          int timer = round(timerFloat);
    String seconds = String(timer);
    String humidity = String(dht.readHumidity());
    String temp = String(dht.readTemperature(true));
    int httpCode = http.POST("name=" + fridgeName + "&seconds=" + seconds + "&temp=" + temp + "&humidity=" + humidity + "&bootup=no");
    String payload = http.getString();
    http.end();
    timerFloat = 0;
    notSent = 0;
    isOpen = 0;
  }
    }
    } else {
      notSent = 0;
      timerFloat = 0;
      isOpen = 0;
    }
  }

// send status

  if (bootUp == 1) {
  if (WiFi.status() == WL_CONNECTED) {
    HTTPClient http;
    http.begin(wifi, statusURL);
    http.addHeader("Content-Type", "application/x-www-form-urlencoded");
          int timer = round(timerFloat);
    String seconds = String(timer);
    String humidity = String(dht.readHumidity());
    String temp = String(dht.readTemperature(true));
    if (isOpen == 0) {
      isOpenString = "no";
    } else {
      isOpenString = "yes";
    };
    int httpCode = http.POST("name=" + fridgeName + "&temp=" + temp + "&humidity=" + humidity + "&isOpen=" + isOpenString + "&bootup=yes");
    String payload = http.getString();
    http.end();
    bootUp = 0;
  }
  }
  statusTimer = statusTimer + 0.1;
  if (statusTimer > statusWaitTime) {
           if (WiFi.status() == WL_CONNECTED) {
    HTTPClient http;
    http.begin(wifi, statusURL);
    http.addHeader("Content-Type", "application/x-www-form-urlencoded");
          int timer = round(timerFloat);
    String seconds = String(timer);
    String humidity = String(dht.readHumidity());
    String temp = String(dht.readTemperature(true));
    if (isOpen == 0) {
      isOpenString = "no";
    } else {
      isOpenString = "yes";
    };
    int httpCode = http.POST("name=" + fridgeName + "&temp=" + temp + "&humidity=" + humidity + "&isOpen=" + isOpenString + "&bootup=no");
    String payload = http.getString();
    http.end();
    statusTimer = 0;
  }
  }
  delay(100);
}
 
 
//-------- Fuctions used for WiFi credentials saving and connecting to it
bool testWifi(void)
{
  int c = 0;
  Serial.println("Waiting for Wifi to connect");
  while ( c < 20 ) {
    if (WiFi.status() == WL_CONNECTED)
    {
      return true;
    }
    delay(500);
    Serial.print("*");
    c++;
  }
  Serial.println("");
  Serial.println("Connect timed out, opening AP");
  return false;
}
 
void launchWeb()
{
  Serial.println("");
  if (WiFi.status() == WL_CONNECTED)
    Serial.println("WiFi connected");
  Serial.print("Local IP: ");
  Serial.println(WiFi.localIP());
  createWebServer();
  // Start the server
  server.begin();
  Serial.println("Server started");
}
 
void setupAP(void)
{
  WiFi.mode(WIFI_STA);
  WiFi.disconnect();
  delay(100);
  int n = WiFi.scanNetworks();
  Serial.println("scan done");
  if (n == 0)
    Serial.println("no networks found");
  else
  {
    Serial.print(n);
    Serial.println(" networks found");
    for (int i = 0; i < n; ++i)
    {
      // Print SSID and RSSI for each network found
      Serial.print(i + 1);
      Serial.print(": ");
      Serial.print(WiFi.SSID(i));
      Serial.print(" (");
      Serial.print(WiFi.RSSI(i));
      Serial.print(")");
      Serial.println((WiFi.encryptionType(i) == ENC_TYPE_NONE) ? " " : "*");
      delay(10);
    }
  }
  Serial.println("");
  st = "<ol>";
  for (int i = 0; i < n; ++i)
  {
    // Print SSID and RSSI for each network found
    st += "<li>";
    st += WiFi.SSID(i);
    st += " (";
    st += WiFi.RSSI(i);
 
    st += ")";
    st += (WiFi.encryptionType(i) == ENC_TYPE_NONE) ? " " : "*";
    st += "</li>";
  }
  st += "</ol>";
  delay(100);
  WiFi.softAP("FridgeGridSetup", "hungerfreedallascounty");
  Serial.println("launching webserver");
  launchWeb();
}
 
void createWebServer()
{
 {
    server.on("/", []() {
 
      IPAddress ip = WiFi.softAPIP();
      String ipStr = String(ip[0]) + '.' + String(ip[1]) + '.' + String(ip[2]) + '.' + String(ip[3]);
      content = "<!DOCTYPE HTML>\r\n<html>Fridge Grid setup page - Hunger Free Dallas County ";
      content += "<form action=\"/scan\" method=\"POST\"><input type=\"submit\" value=\"scan\"></form>";
      content += ipStr;
      content += "<p>";
      content += st;
      content += "</p><form method='get' action='setting'><label>SSID: </label><input name='ssid' length=32><br><label>Password: </label><input name='pass' length=32><br><label>Fridge Name: </label><input name='fridgename' length=32><br><input type='submit'></form>";
      content += "</html>";
      server.send(200, "text/html", content);
    });
    server.on("/scan", []() {
      //setupAP();
      IPAddress ip = WiFi.softAPIP();
      String ipStr = String(ip[0]) + '.' + String(ip[1]) + '.' + String(ip[2]) + '.' + String(ip[3]);
 
      content = "<!DOCTYPE HTML>\r\n<html>go back";
      server.send(200, "text/html", content);
    });
 
    server.on("/setting", []() {
      String qsid = server.arg("ssid");
      String qpass = server.arg("pass");
      String qname = server.arg("fridgename");
      if (qsid.length() > 0 && qpass.length() > 0) {
        Serial.println("clearing eeprom");
        for (int i = 0; i < 96; ++i) {
          EEPROM.write(i, 0);
        }
        Serial.println(qsid);
        Serial.println("");
        Serial.println(qpass);
        Serial.println("");
        Serial.println(qname);
        Serial.println("");
 
        Serial.println("writing eeprom ssid:");
        for (int i = 0; i < qsid.length(); ++i)
        {
          EEPROM.write(i, qsid[i]);
        }
        Serial.println("writing eeprom pass:");
        for (int i = 0; i < qpass.length(); ++i)
        {
          EEPROM.write(32 + i, qpass[i]);
        }
        Serial.println("writing eeprom fridge name:");
        for (int i = 0; i < qname.length(); ++i)
        {
          EEPROM.write(64 + i, qname[i]);
        }
        EEPROM.commit();
 
        content = "{\"Success\":\"saved to eeprom... reset to boot into new wifi\"}";
        statusCode = 200;
        ESP.reset();
      } else {
        content = "{\"Error\":\"404 not found\"}";
        statusCode = 404;
        Serial.println("Sending 404");
      }
    });
  } 
}