function calculateAverage(array) {
    var total = 0;
    var count = 0;

    array.forEach(function(item, index) {
        total += item;
        count++;
    });

    return total / count;
}
function parseTime(time) {
  const hour = parseInt(time.split(":")[0])
  const minute = parseInt(time.split(":")[1])
  if (hour > 12) {
     newHour = hour - 12
     amOrPM = "PM"
  } else {
     if (hour == 0) {
        newHour = 12
        amOrPM = "AM"
     } else {
        newHour = hour
        amOrPM = "AM"
     }
  }
  console.log(minute)
  if (minute < 10) {
     newMinute = "0" + minute
  } else {
     newMinute = minute
  }
  return newHour + ":" + newMinute + " " + amOrPM
}
function csvToArray(str, delimiter = ",") {
  const headers = str.slice(0, str.indexOf("\n")).split(delimiter);
  const rows = str.slice(str.indexOf("\n") + 1).split("\n");
  const arr = rows.map(function (row) {
    const values = row.split(delimiter);
    const el = headers.reduce(function (object, header, index) {
      object[header] = values[index];
      return object;
    }, {});
    return el;
  });
  return arr.slice(0, -1)
}
  let xhr = new XMLHttpRequest();
  xhr.open("GET", "/api/get_json");
  xhr.setRequestHeader("Content-Type", "application/json");
  xhr.setRequestHeader("Cache-Control", "no-cache, no-store, max-age=0");
  xhr.responseType = 'json';
  xhr.send();
  xhr.onload = function() {
    var listResponse = xhr.response
    var listNum = Object.keys(listResponse).length
    console.log(listNum)
    var select = document.createElement("select");
    select.name = "grids";
    select.id = "grids"
    for (const name in listResponse)
    {
        var option = document.createElement("option");
        option.value = listResponse[name];
        option.text = listResponse[name]
        select.appendChild(option);
    }
    var label = document.createElement("label");
    label.innerHTML = "Choose your fridge: "
    label.htmlFor = "grids";
    document.getElementById("dropdown").appendChild(label).appendChild(select);
  }

function submitEmail() {
  var select = document.getElementById('grids');
  selectedFridge = select.options[select.selectedIndex].value;
  email = document.getElementById("email").value
  let xhr = new XMLHttpRequest();
  xhr.open("GET", "/api/save_email?name=" + selectedFridge + "&email=" + email);
  xhr.setRequestHeader("Content-Type", "application/json");
  xhr.setRequestHeader("Cache-Control", "no-cache, no-store, max-age=0");
  xhr.send();
  xhr.onload = function() {
    resp = JSON.stringify(xhr.response)
    if (resp.includes("ok")) {
      alert("Success!")
    } else {
      alert(resp)
    }
  }

}
function goToFridge() {
  var select = document.getElementById('grids');
  selectedValue = select.options[select.selectedIndex].value;
  getCSVName = selectedValue.replace(/ /g,"_") + ".csv";
  let xhr = new XMLHttpRequest();
  xhr.open("GET", "/openLogs/" + getCSVName);
  xhr.setRequestHeader("Content-Type", "application/json");
  xhr.setRequestHeader("Cache-Control", "no-cache, no-store, max-age=0");
 // xhr.responseType = 'json';
  xhr.send();
  xhr.onload = function() {
   console.log(xhr.response)
     const fridgeOpenLog = csvToArray(xhr.response);
     console.log(fridgeOpenLog)
     var logSecondsArray = fridgeOpenLog.map(function(i) {
  console.log(i)
  return parseInt(i.Seconds);
});
     var logHoursArray = fridgeOpenLog.map(function(i) {
  console.log(i)
  return i.Time;
});
console.log("Hours: " + logHoursArray)
console.log(logSecondsArray)
var logAmount = Object.keys(logSecondsArray).length
//var logSeconds = logSecondsArray.reduce((a, b) => parseInt(a) + parseInt(b), 0)
var logSeconds = logSecondsArray.reduce((a, b) => a + b, 0)
console.log(logAmount)
console.log(logSeconds)
var lastOpen = fridgeOpenLog[fridgeOpenLog.length - 1];
    stats = document.getElementById('stats');
    const nameH2 = document.createElement('h1');
    const openedH2 = document.createElement('h2');
    const secondsH2 = document.createElement('h2');
    const lastOpenH2 = document.createElement('h2');
    const avgSecsH2 = document.createElement('h2');
    const lastTempH2 = document.createElement('h2');
    const lastHumidityH2 = document.createElement('h2');
    nameH2.textContent = selectedValue
    openedH2.textContent = "Times Opened: " + `${logAmount}`
    secondsH2.textContent = "Seconds Open: " + `${logSeconds}`
    avgSecsH2.textContent = "Average Seconds per Open: " + `${calculateAverage(logSecondsArray)}`
    lastOpenH2.textContent = "Last Open: " + lastOpen.Date + " at " + parseTime(lastOpen.Time)
    stats.innerHTML = ''
    stats.appendChild(nameH2);
    stats.appendChild(openedH2);
    stats.appendChild(secondsH2)
    stats.appendChild(avgSecsH2)
    stats.appendChild(lastOpenH2)
//    stats.appendChild(document.createElement('h2').textContent = "Last Reported Temperature: " + lastOpen.Temp + " F")
//    stats.appendChild(document.createElement('hr'))
  var select = document.getElementById('grids');
  var value = select.options[select.selectedIndex].value;
  getCSVName = value.replace(/ /g,"_") + ".csv";
  let xhr2 = new XMLHttpRequest();
  xhr2.open("GET", "/statusLogs/" + getCSVName);
  xhr2.setRequestHeader("Content-Type", "application/json");
  xhr2.setRequestHeader("Cache-Control", "no-cache, no-store, max-age=0");
 // xhr2.responseType = 'json';
  xhr2.send();
  xhr2.onload = function() {
const fridgeStatusLog = csvToArray(xhr2.response);
var lastStatus = fridgeStatusLog[fridgeStatusLog.length - 1];
console.log(lastStatus)
console.log(lastStatus.Time)
const lastUpdateH2 = document.createElement('h2');
lastTempH2.textContent = "Last Reported Temperature: " + lastStatus.Temp + " F"
lastHumidityH2.textContent = "Last Reported Humidity: " + lastStatus.Humidity + "%"
lastUpdateH2.textContent = "Last Status Update: " + lastStatus.Date + " at " + parseTime(lastStatus.Time)
stats.appendChild(lastUpdateH2)
stats.appendChild(lastTempH2)
stats.appendChild(lastHumidityH2)
//    stats.appendChild(document.createElement('hr'))
console.log(parseTime(lastStatus.Time))
statsChart = document.getElementById('secondsHourChart');
statsChart.innerHTML = ''
new Chart("secondsHourChart", {

  type: "line",
  data: {
    labels: logHoursArray,
    datasets: [{
      label: "Seconds",
      backgroundColor: "rgba(0,0,0,1.0)",
      borderColor: "rgba(0,0,0,0.1)",
      data: logSecondsArray
    }]
  },
  options:{   
responsive: true,
 plugins: {
      title: {
        display: true,
        text: 'Chart.js Line Chart'
      },
    },
    scales: {
      y: {
        display: true,
        title: {
          display: true,
          text: 'Month'
        }
      }
    }}
});
endDiv = document.getElementById('lastDiv');
endDiv.innerHTML = ''
var a1 = document.createElement('a');
var link1 = document.createTextNode("Raw Status CSV (updated every ~5 minutes, does not contain seconds)");
a1.appendChild(link1)
a1.title = "Raw Status CSV";
a1.href = "/statusLogs/" + getCSVName;
endDiv.appendChild(a1)
var a2 = document.createElement('a');
var link2 = document.createTextNode("Raw Opens CSV (actual data, updated on every fridge open)");
a2.appendChild(link2)
a2.title = "Raw Opens CSV";
a2.href = "/openLogs/" + getCSVName;
endDiv.appendChild(document.createElement('br'));
endDiv.appendChild(a2)
//endDiv.innerHTML = ''
    endDiv.appendChild(document.createElement('hr'))
  }
  }
}
