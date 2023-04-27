var form = new FormData();
var logs = [];
var optExt = "/extractaudio";

var host = "http://api.bean-bon.eu";

var custom_placeholder = [{
        "Que": "videoaudioextractQ",
        "Name": "Extract audio",
        "Output": "track.mp3"
    },
    {
        "Que": "transcribeQ",
        "Name": "Transcribe",
        "Params": {
            "model": "base",
            "language": "en"
        }
}]

var lastObj;



function readURL(input) {
    if (input.files && input.files[0]) {
        displaySwap(true);
        $('#input-video')[0].src = URL.createObjectURL(input.files[0]);
        $('#input-video').parent()[0].load();
    }
}

const getOpt = (selection) => {
    if (selection.value != "custom") {
        $('#custom-opts').height(0);        
        setTimeout((val) => {
            val.hide();
        }, 200, $('#custom-opts'));

    }
    switch (selection.value) {
        case "audio":
            optExt = "/extractaudio";
            break;
        case "audiot":
            optExt = "/transcribe";
            break;
        case "custom":
            optExt = "/custom";
            $('#custom-opts').show();
            $('#custom-opts').height(200);
            $('#custom-opts').val(JSON.stringify(custom_placeholder,null, 2));
            break;
        default:
            optExt = "/extractaudio";
            break;
    }
}

// FIX
const sendRequest = () => {
    form = new FormData();
    form.append("file", $('#file-input')[0].files[0]);
    // Check if custom and append data
    if ($('#action-select').find(":selected").val() == "custom")
        form.append("routing", $('#custom-opts').val());

    var settings = {
        "url": host + optExt,
        "method": "POST",
        "timeout": 0,
        "processData": false,
        "mimeType": "multipart/form-data",
        "contentType": false,
        "data": form
    };
    newLog("Requesting process", host + optExt)
    loadingText("Requesting new process.")
    loadingBar(0);
    $('#loading-spinner').show();
    $('#loading').show();
    $('#output').hide();
    $.ajax(settings).done(function (response) {
        let info = JSON.parse(response);
        consLog(info);
        newLog("Created bucket", JSON.stringify(info));
        newLog("Tracking bucket changes", info.bucket);
        trackStatus(info.bucket);
    }).fail(function (response) {
        responseError(response)
    });
}

const responseError = (error) => {
    $('#loading-spinner').hide();
    loadingBar(0);
    loadingText("Error! Response " + error.status + " when requesting")
    console.log(error);
}

const consLog = (obj) => {
    if (!_.isEqual(lastObj, obj)) {
        lastObj = obj;
        console.log(obj);
    }
}

const trackStatus = (bucket) => {
    var settings = {
        "url": host + '/sf/' + bucket + '/info',
        "method": "GET",
        "timeout": 0,
    };

    $.ajax(settings).done(function (response, status) {
        consLog(response);
        // Loading progress log
        var text = "";
        if (response.Routing[response.Stage].Name != null){
            text = "Stage " + (response.Stage + 1) + "/" + response.Routing.length + ": " + response.Routing[response.Stage].Name + " - " + response.Status
        } else {
            text = "Stage " + (response.Stage + 1) + "/" + response.Routing.length + ": " + response.Routing[response.Stage].Que + " - " + response.Status
        }
        loadingText(text);
        
        // loading bar progress change
        var bonus = 0;
        if (response.Status != "waiting") {
            bonus = 1 / response.Routing.length / 2;
        }
        loadingBar(response.Stage / response.Routing.length + bonus);

        // checks
        persistantLog("Checking for changes", response.Status);
        if (!(response.Status == "Completed" || response.Status == "error")){
            setTimeout((buck) => {trackStatus(buck)}, 200, bucket);
        } else if (response.Status == "Completed"){
            completeRequest(response.ID);
        } else if (response.Status == "error"){
            console.log(status, response)
        }
    }).fail(function (response) {
        responseError(response)
    });
}

const loadingText = (text) => {
    $('#loading-text').get(0).lastChild.nodeValue = "  " + text;
}

const loadingBar = (progress) => {
    $('#loading-progress').width((progress * 100) + "%")
}

const completeRequest = (bucket) => {
    // Loading stuff
    loadingText("Completed - Downloading result");
    loadingBar(1);
    $('#loading-spinner').hide();

    newLog("Downloading result", host + '/sf/' + bucket);

    var download = host + '/sf/' + bucket
    $('#output').show();
    $('#output-download').attr("href", download);
}

const newLog = (log, extra) => {
    logs.push([Date().valueOf().substr(16,8), log, extra]);
    printLog();
}

const persistantLog = (log, extra) => {
    var i = logs.length-1;
    if (logs[i][1] != log || logs[i][2] != extra)
        newLog(log, extra)
    else {
        if (logs[i].length < 4)
            logs[i].push("x1")
        else
            logs[i][3] = "x" + (parseInt(logs[i][3].substr(1)) + 1)
    }
    printLog();
}


const printLog = () => {
    var text = "";
    for (const line of logs) {
        for (const item of line) {
            text += item + ' | ';
        }
        text += "\n";
    }
    $('#log-block').val(text);
    $('#log-block').scrollTop($('#log-block')[0].scrollHeight);
}

const reset = () => {
    displaySwap(false);
    $('#file-input').val('');
    $('#loading').hide();
    $('#output').hide();
    $('#log-block').val('');
}

const displaySwap = (main) => {
    if (main) {
        $('#file-picker').hide();
        $('#main').show();
    } else {
        $('#file-picker').show();
        $('#main').hide();
    }
}

const toggleLogs = () => {
    const logblock = $('#log-block');
    const logbutton = $('#log-toggle');
    if (logblock.height() == 0) {
        logbutton.get(0).firstChild.nodeValue = "Hide Logs ";
        $('#log-toggle-fa').toggleClass('fa-arrow-down fa-arrow-up-from-bracket')
        logblock.show();
        logblock.height(200);
    } else {
        logbutton.get(0).firstChild.nodeValue = "Show logs ";
        $('#log-toggle-fa').toggleClass('fa-arrow-down fa-arrow-up-from-bracket')
        logblock.height(0);
        setTimeout((val) => {
            val.hide();
        }, 200, logblock);
    }
}