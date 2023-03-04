const STOP = 2,
    SUCCESS = 0,
    FAILED = 1,
    SHT_TYPE_D_NOMASIC = 684,
    SHT_TYPE_D_ANCHOR = 685,
    SHT_TYPE_A_FC2 = 368,
    SHT_TYPE_A_MCRACK = 672,
    SHT_TYPE_A_LEAKED = 654;
var $videoPlayControl = `<video src="/static/video/青梅竹马.mp4" controls id="video-player"></video>`,
    module_default = "home",
    category_default = 0,
    videoPageLimit = 18,
    playVideoId,
    player,
    screenshootIdx = 0,
    screenhootMapFile = new Map(),
    spiderLimit = 10,
    spiderPageThreshold = 10;;

function ajaxFunc(url, type, data, success) {
    $.ajax({
        url: url,
        type: type,
        dataType: "json",
        data: data,
        error: function () {

        },
        success: success,
    })
}

function showTipToast(msg, time, mode) {
    if (time == undefined) {
        time = 5
    }
    $(`#tip-toast`).attr('data-delay', time * 1000)
    $(`#tip-toast .toast-body`).html(msg)
    $('#tip-toast').toast('show')
}

$(function () {
    // set nav item active
    try {
        var $activeNote = $('#active-note'),
            activeModule = $activeNote.attr('active-module'),
            activeCategory = $activeNote.attr('active-category');
        if (activeModule == "" || activeModule == undefined) activeModule = module_default;
        if (activeCategory == "") activeCategory = category_default;
        $("#top-nav .nav-item.active").removeClass("active")
        $("#top-nav .nav-item[data-n=" + activeModule + "]").addClass("active")
        $("#lside-nav .subnav-item.active").removeClass("active")
        $("#lside-nav .subnav-item[data-i=" + activeCategory + "]").addClass("active")
    } catch { }
    try {
        bsCustomFileInput.init()
    } catch { }
    window.addEventListener('load', function () {
        // Fetch all the forms we want to apply custom Bootstrap validation styles to
        var forms = document.getElementsByClassName('needs-validation');
        // Loop over them and prevent submission
        var validation = Array.prototype.filter.call(forms, function (form) {
            form.addEventListener('submit', function (event) {
                if (form.checkValidity() === false) {
                    event.preventDefault();
                    event.stopPropagation();
                }
                form.classList.add('was-validated');
            }, false);
        });
    }, false);

    $('#tag-add-input').on('input propertychange', function () {
        let text = $(this).val(),
            $tags = $('#tag-select-list-tag .tag-item'),
            len = $tags.length;
        $tags.removeClass('hidden')
        for (let i = 0; i < len; i++) {
            if ($tags.eq(i).html().indexOf(text) == -1) {
                $tags.eq(i).addClass('hidden')
            }
        }
    })

    $('#actor-add-input').on('input propertychange', function () {
        let text = $(this).val(),
            $tags = $('#tag-select-list-actor .tag-item'),
            len = $tags.length;
        $tags.removeClass('hidden')
        for (let i = 0; i < len; i++) {
            if ($tags.eq(i).html().indexOf(text) == -1) {
                $tags.eq(i).addClass('hidden')
            }
        }
    })

    $('#actor-filter-input').on('input propertychange', function () {
        let text = $(this).val(),
            $tags = $('#side-list-actor .tag-item'),
            len = $tags.length;
        $tags.removeClass('hidden')
        for (let i = 0; i < len; i++) {
            if ($tags.eq(i).html().indexOf(text) == -1) {
                $tags.eq(i).addClass('hidden')
            }
        }
    })

    $('.to-down-kit').on('click', function () {
        let $elem = $($(this).attr('data-target') + '');
        $elem.scrollTop($elem.prop("scrollHeight"))
    })
    $('.to-top-kit').on('click', function () {
        let $elem = $($(this).attr('data-target') + '');
        $elem.scrollTop(-$elem.prop("scrollHeight"))
    })

    $('#video-player-modal').on('hide.bs.modal', function (event) {
        try {
            player.pause()
        } catch { }
        $('#video-player-modal').removeClass('large-modal')
    })

    $('.dynamic-select').on('change', function () {
        let $aims = $(`#${$(this).attr('data-f')} option`),
            value = $(this).val();
        if (value == "") {
            $aims.removeClass('hidden');
        } else {
            for (let i = 1, len = $aims.length; i < len; i++) {
                if ($aims.eq(i).attr('data-f') == value)
                    $aims.eq(i).removeClass('hidden');
                else $aims.eq(i).addClass('hidden');
            }
        }
    })

    $('.click-view-pic').click(function () {
        console.log($(this).attr('src'))
        $('#picture-view-modal img').attr('src', $(this).attr('src'))
        $('#picture-view-modal').modal('show')
    })

    $('#record-page .category-item').click(function () {
        $('#record-page [name=record-show]').attr("checked", false)
        if ($(this).hasClass('active')) {
            return
        }
        $('#record-page .category-item').removeClass('active')
        $(this).addClass('active')
        let category = $(this).attr('data-i');
        ajaxFunc("/record.json", "GET", { id: category, action: "category" }, function (res) {
            if (res.code != 0) showTipToast(res.msg)
            else setRecordList(res)
        })
    })
})

/****************************************
 *                                      *
 *               VIDEO                  *
 *                                      *
*****************************************/

function getVideoRecommendList(aim) {
    let nowdate = new Date().toLocaleDateString().replaceAll("/", "-"),
        dataDate = $('#video-recommend-list').attr('data-d'),
        date = nowdate;
    if (dataDate == "") {
        dataDate = nowdate;
    }
    if (dataDate == nowdate && aim == "post") {
        return
    }
    let t = new Date(dataDate);
    switch (aim) {
        case 'pre':
            t.setDate(t.getDate() - 1)
            date = t.toLocaleDateString().replaceAll("/", "-");
            break;
        case 'post':
            t.setDate(t.getDate() + 1)
            date = t.toLocaleDateString().replaceAll("/", "-");
            break;
        case 'refresh':
            break;
    }
    dataDate = date
    let strarr = date.split('-');
    for (let i = 1; i < 3; i++) {
        let num = parseInt(strarr[i])
        if (num <= 9) strarr[i] = '0' + num;
    }
    date = strarr[0] + '-' + strarr[1] + '-' + strarr[2]
    ajaxFunc(`video.json`, "POST", { action: "recommendlist", aim: aim, date: date }, function (res) {
        if (res.code == 0) {
            let $list = $('#video-recommend-list .video-list');
            $list.empty();
            for (let v of res.data) {
                let tagstr = '';
                if (v.actors != null) {
                    for (let t of v.actors) {
                        tagstr += `<a href="/actor?id=${t.id}" class="tag-item">${t.name}</a>`
                    }
                }
                if (v.tags != null) {
                    for (let t of v.tags) {
                        tagstr += `<a href="/tag?id=${t.id}" class="tag-item">${t.name}</a>`
                    }
                }
                $list.append(`
                    <div class="video-item">
                        <a onclick="playVideoByModal(${v.id})" class="video-cover">
                            <img src="${v.cover}" lazyload="on" alt="">
                            <span class="video-duration">${v.duration}</span>
                            <span class="video-category">${v.actegorytitle}</span>
                        </a>
                        <div class="video-item-info">
                            <a onclick="playVideoByModal(${v.id})" class="video-title" alt="${v.title}">${v.title}</a>
                            <div class="video-details">
                                <i class="bi bi-eye"></i> ${v.view}&ensp;&ensp;<i class="bi bi-heart"></i>
                                ${v.collect}&emsp;&emsp;<i class="bi bi-clock"></i>
                                ${v.pubtime}
                            </div>
                            <div class="video-tags">
                                ${tagstr}
                            </div>
                        </div>
                    </div>
                `)
            }
        }
        if (aim != "refresh") {
            $('#video-recommend-list').attr('data-d', dataDate);
            $('#video-recommend-list .list-date').html(date);
        }
    })
}

var configVideo = function (path, cover, id, timenodes) {
    if (id == playVideoId) {
        return
    }
    if (player != undefined) player.dispose();
    var captionStr = `<track kind="captions" src="${path.replace("mp4", "vtt")}" srclang="zh" label="中文" default>`
    $("#video-player-modal .video-player-container").html(`<video id="video-player" class="video-js vjs-default-skin vjs-matrix vjs-big-play-centered  vjs-16-9 theater-mode" controls preload="auto">${captionStr}</video>`);
    player = videojs('video-player', {
        controls: true,
        loop: false,
        poster: cover.replaceAll('\\', "\\\\"),
        fluid: true,
        // width: '960px',
        // height: '540px',
        liveui: true,
        sources: [{
            src: path,
            type: 'video/mp4'
        }],
        notSupportedMessage: '此视频暂无法播放!', // 无法播放时显示的信息
        playbackRates: [0.5, 1, 2, 4],
        controlBar: {
            progressControl: true, // 进度条
            currentTimeDisplay: true, // 当前时间
            timeDivider: true, // 时间分割线
            durationDisplay: true, // 总时间
            remainingTimeDisplay: false, //剩余时间
            customControlSpacer: true,
            playbackRateMenuButton: true,
            fullscreenToggle: true,         // 全屏按钮
        },
        plugins: {},
    }, function onPlayerReady() {
        var firstPlay = true;
        this.on('play', function () {//开始播放
            playVideoId = id;
            if (firstPlay) {
                if (timenodes != undefined) {
                    for (let node of timenodes) {
                        var leftPos = node * 100 / this.duration() + '%';
                        $('.vjs-progress-holder').prepend(`<div class="time-node-item bi bi-geo-alt-fill" onclick="videoPlayJumpTo(${node})" style="left: ${leftPos};"></div>`);
                    }
                }
                ajaxFunc('/video.json', 'POST', { id: id, action: 'play' })
                firstPlay = false
            }
        });
    });
    videojs('video-player').ready(function () {
        this.hotkeys({
            volumeStep: 0.05,
            seekStep: 5,
            enableVolumeScroll: false, //禁用鼠标滚轮调节问音量大小
            enableModifiersForNumbers: false,
        })
    })
}


function videoPlayJumpTo(time) {
    player.currentTime(time)
}

// play video
function playVideoByModal(id) {
    $("#video-player-modal #video-relate-list").html('');
    $('.screenshoot-list').html('');
    // $('#video-player-modal').removeClass('large-modal');
    ajaxFunc("/video.json", "GET", { id: id, action: "videoplay" }, function (res) {
        if (res.code != 0) {
            alert(res.msg)
        } else {
            // $("#video-player-modal .video-player-container").html(`<video src="${res.data.path}" controls id="video-player"></video>`)
            $("#video-play-id").attr("data-i", res.data.id)
            $("#video-player-modal .modal-title").text(res.data.title)
            configVideo(res.data.path, res.data.cover, res.data.id, res.data.timenodes)
            for (let v of res.data.relateVideos) {
                $("#video-player-modal #video-relate-list").append(
                    `<a onclick="playVideoByModal(${v.id})" class="video-item"><div class="video-cover"><img src="${v.cover}"><div class="video-title">${v.title}</div><div class="video-duration">${v.duration}</div></div></a>`)
            }
            for (let sc of res.data.screenshoots) {
                screenshootIdx++
                $('.screenshoot-list').append(`<div class="screenshoot-item" data-i="${screenshootIdx}"><img class="click-view-pic" src="${sc.path}" alt=""></div>`)
                $('.click-view-pic').click(function () {
                    $('#picture-view-modal img').attr('src', $(this).attr('src'))
                    $('#picture-view-modal').modal('show')
                })
            }
        }
    })
    $('#video-player-modal').modal('show');
}


function toggleVideoModal() {
    if ($('#video-player-modal').hasClass('large-modal')) {
        $('#video-player-modal').removeClass('large-modal');
    } else {
        $('#video-player-modal').addClass('large-modal');
    }
}

function editVideoByModal(id) {
    // get video infomation
    ajaxFunc("/video.json", "GET", { id: id }, function (res) {
        if (res.code != 0) {
            alert(res.msg)
        } else {
            setVideoEidtModal(res.data)
            $('#video-edit-modal').modal('show')
        }
    })
}

function setVideoEidtModal(obj) {
    document.getElementById('videoinfo-form').reset();
    $('#video-edit-modal .video-info-id').text(obj.id);
    $('#video-edit-modal input[name=id]').val(obj.id);
    $('#video-edit-modal input[name=title]').val(obj.title);
    var title = obj.title;
    title = title.toLowerCase();
    if (title.startsWith('tokyo')) {
        title = title.replace('tokyo hot ', '')
        title = title.replace('tokyo-hot ', '')
    }
    if (title.startsWith("fc2")) {
        title = title.replace('fc2 ', 'fc2-')
        title = title.replace('fc2-ppv ', 'fc2-ppv-')
    }
    $('#video-edit-modal #spider-title').val(title);
    $('#video-edit-modal input[name=path]').val(obj.path);
    $('#video-edit-modal input[value=' + obj.categoryid + ']').attr('checked', 'true');
    $('#video-edit-modal .video-cover img').attr("src", obj.cover);
    // 初始化 tag select list
    $('#tag-add-input').val('')
    $(`#tag-select-list-tag .tag-item.hidden`).removeClass('hidden');
    $(`#tag-select-list-tag .tag-item.tag-select`).removeClass('tag-select');
    $('#actor-add-input').val('')
    $(`#tag-select-list-actor .tag-item.hidden`).removeClass('hidden');
    $(`#tag-select-list-actor .tag-item.tag-select`).removeClass('tag-select');

    let $tagActorInput = $('#tag-input-actor'),
        $tagTagInput = $('#tag-input-tag');
    $tagActorInput.html('');
    $tagTagInput.html('');
    // 配置tag select list 和 tag input
    if (obj.actorid != "") {
        let actorList = JSON.parse(`${obj.actorid}`);
        for (let elem of actorList) {
            $(`#tag-select-list-actor .tag-item[data-i=${elem}]`).addClass('tag-select');
            $tagActorInput.append(getTagInputElem(elem, 'actor'));
        }
    }
    if (obj.tagid != "") {
        let tagList = JSON.parse(`${obj.tagid}`);
        for (let elem of tagList) {
            $(`#tag-select-list-tag .tag-item[data-i=${elem}]`).addClass('tag-select');
            $tagTagInput.append(getTagInputElem(elem, 'tag'));
        }
    }
}

function videoinfoEditSubmit() {
    let $tagsSelect = $('#tag-select-list-tag .tag-item.tag-select'),
        $actorssSelect = $('#tag-select-list-actor .tag-item.tag-select'),
        tagArr = new Array(),
        actorArr = new Array();
    for (let i = 0, len = $tagsSelect.length; i < len; i++) {
        tagArr.push($tagsSelect.eq(i).attr('data-i') + '');
    }
    for (let i = 0, len = $actorssSelect.length; i < len; i++) {
        actorArr.push($actorssSelect.eq(i).attr('data-i') + '');
    }

    $('#videoinfo-form input[name=tagid]').val(JSON.stringify(tagArr))
    $('#videoinfo-form input[name=actorid]').val(JSON.stringify(actorArr))

    let data = $('#videoinfo-form').serialize();

    ajaxFunc('/video.json', 'POST', data, function (res) {
        showTipToast(`修改视频[ ${$('#video-edit-modal input[name=id]').val()} ]内容成功！`)
    })
}

function videoEditZoom(offset) {
    let $vlist = $('.video-list .video-item'),
        nowIdx = $('#video-edit-modal input[name=id]').val(),
        idx = 0;
    for (let i = 0; i < videoPageLimit; i++) {
        if ($vlist.eq(i).attr('data-i') == nowIdx) {
            idx = i + offset
            break
        }
    }
    if (idx < 0 || idx >= videoPageLimit) {
        showTipToast('当前是第一个/最后一个！')
        return
    }
    editVideoByModal($vlist.eq(idx).attr('data-i'))
}

function selecttagTag(id) {
    selectTag(id, 'tag');
}

function selectactorTag(id) {
    selectTag(id, 'actor');
}

function selectTag(id, obj) {
    let $tagSelect = $(`#tag-select-list-${obj} .tag-item[data-i=${id}]`)
    if ($tagSelect.hasClass('tag-select')) {
        return
    }
    $('#tag-input-' + obj).append(getTagInputElem(id, obj))
    $tagSelect.addClass('tag-select');
}

function getTagInputElem(id, obj) {
    let title = $(`#tag-select-list-${obj} .tag-item[data-i=${id}]`).eq(0).html();
    return `<div class="tag-item" data-i="${id}"><span>${title}</span><span class="tag-item-close" onclick="exclude${obj}Tag(${id})">&times;</span></div>`
}

function getTagSelectElem(id, name, obj) {
    return `<div class="tag-item" onclick="select${obj}Tag(${id})" data-i="${id}">${name}</div>`
}

function excludetagTag(id) {
    excludeTag(id, 'tag');
}
function excludeactorTag(id) {
    excludeTag(id, 'actor');
}

function excludeTag(id, obj) {
    $(`#tag-input-${obj} .tag-item[data-i=${id}]`).remove();
    $(`#tag-select-list-${obj} .tag-item[data-i=${id}]`).removeClass('tag-select');
}

function addTag(obj) {
    let name = "";
    switch (obj) {
        case "tag":
            name = $("#tag-add-input").val();
            break;
        case "actor":
            name = $("#actor-add-input").val();
            break;
    }
    if (name == "") {
        showTipToast("当前添加内容为空！")
        return
    }
    ajaxFunc('/tag.json', 'POST', {
        name: name,
        obj: obj,
        action: 'add',
    }, function (res) {
        if (res.code == 0) {
            showTipToast(`添加标签<span class="bg-info">${name}</span>成功！`)
            let tag = res.data;
            $(`#tag-select-list-${obj}`).prepend(getTagSelectElem(tag.id, tag.name, obj))
        } else {
            showTipToast(res.msg)
        }
    })
}

function deleteVideo(id) {
    if (confirm('是否删除当前视频？')) {
        if (player != undefined) {
            player.dispose();
            playVideoId = undefined;
        }
        ajaxFunc(`/video.json`, "POST", { id: id, action: "delete" }, function (res) {
            if (res.code == 0) {
                showTipToast("视频删除成功！")
                location.reload()
            } else {
                showTipToast("视频删除失败！")
            }
        })
    }
}

function collectVideo() {
    let id = $("#video-play-id").attr("data-i");
    if (id == "" || id == undefined) {
        return
    }
    ajaxFunc(`/video.json`, "POST", { id: id, action: "collect" }, function (res) {
        showTipToast(res.msg)
    })
}

function uploadLocalCover() {
    let id = $('#video-edit-modal input[name=id]').val(),
        files = $("#localFile").prop('files'),
        formdata = new FormData();
    if (id == undefined || id == "" || files.length == 0) {
        showTipToast(`<span class="bg-warn">当前视频表单所需内容为空</span>`)
        return
    }
    formdata.append("file", files[0]);
    formdata.append("action", "cover");
    formdata.append("id", id);
    $.ajax({
        url: "/video.json"
        , type: "POST"
        , data: formdata
        , processData: false // 告诉jQuery不要去处理发送的数据
        , contentType: false // 告诉jQuery不要去设置Content-Type请求头
        , async: false
        , dataType: "json"
        , success: function (res) {
            showTipToast(res.msg)
            if (res.code == 0) {
                $(`#video-edit-modal .video-cover img`).attr('src', res.data);
                $(`.video-list .video-item[data-i=${id}] .video-cover img`).attr('src', res.data);
            }
        }
    })
}

function spiderCover() {
    var title = $('#video-edit-modal #spider-title').val(),
        baseurl = $('#video-edit-modal #spider-base-url').val(),
        baseIdx = baseurl.lastIndexOf('?'),
        id = $('#video-edit-modal input[name=id]').val(),
        spiderTarget = baseurl.substr(baseIdx + 1),
        action = "cover";
    let url = baseurl.substr(0, baseIdx) + title;
    if (spiderTarget == "web") {
        url = title;
        action = "coverdownload"
    }
    if (spiderTarget == "javdoe" || spiderTarget == "javporn") {
        url = url + "/"
    }
    console.log('spider cover from url:', url)
    ajaxFunc('/spidervinfo.json', 'GET', {
        id: id,
        url: url,
        action: action,
        target: spiderTarget,
    }, function (res) {
        if (res.code == 0) {
            showTipToast('爬取信息成功！')
            // $('#video-edit-modal input[name=cover_network]').val(res.data)
            $('#video-edit-modal .video-cover img').attr('src', res.data)
            $(`.video-list .video-item[data-i=${id}] .video-cover img`).attr('src', res.data);
        } else {
            showTipToast(res.msg)
        }
    })
}

function VideoCoverSpider() {
    var $msgList = $('#video-spider-pane .msg-list').eq(0),
        $spiderAllCount = $('.spider-all-count').eq(0),
        $spiderSuccessCount = $('.spider-success-count').eq(0),
        $spiderFailCount = $('.spider-fail-count').eq(0);

    socket = new WebSocket('ws://' + window.location.host + '/spidervideocover');
    // Message received on the socket
    socket.onmessage = function (event) {
        var data = JSON.parse(event.data);
        $spiderAllCount.html(parseInt($spiderAllCount.html()) + 1)
        $msgList.append(`<div class="msg-item"><span class="msg-time">${data.time}</span><div class="msg-content">${data.msg}</div></div>`)
        $msgList.scrollTop($msgList.prop("scrollHeight"))
        switch (data.code) {
            case 0: // SUCCESS
                $spiderSuccessCount.html(parseInt($spiderSuccessCount.html()) + 1)
                break;
            case 1: // FAILED
                $spiderFailCount.html(parseInt($spiderFailCount.html()) + 1)
                break;
            case 2: // MESSAGE
                break;
        }
    };
}

function shootVideo() {
    const video = document.querySelector('#video-player_html5_api'),
        scale = 1;
    var time = player.currentTime(),
        id = $("#video-play-id").attr("data-i"),
        vname = $('#video-player-modal .modal-title').html();
    var canvas = document.createElement("canvas");
    canvas.width = video.videoWidth * scale;
    canvas.height = video.videoHeight * scale;
    canvas.getContext('2d').drawImage(
        video,
        0, 0,
        canvas.width, canvas.height
    );
    var img = document.createElement("img");
    img.src = canvas.toDataURL('image/png');

    let arr = img.src.split(","),
        mime = arr[0].match(/:(.*?);/)[1],
        bstr = atob(arr[1]),
        n = bstr.length,
        u8arr = new Uint8Array(n);
    while (n--) {
        u8arr[n] = bstr.charCodeAt(n);
    }
    let blob = new Blob([u8arr], { type: mime });
    screenshootIdx++;
    const file = new File([blob], `[${id}][${vname}]${time}.png`, { type: mime });
    screenhootMapFile[screenshootIdx] = file;
    $('.screenshoot-list').append(
        `<div class="screenshoot-item" data-i="${screenshootIdx}"><img class="click-view-pic" src="${URL.createObjectURL(blob)}" alt="">
            <div class="kit-list">
                <div class="kit-item" onclick="deleteScreenshoot(${screenshootIdx})"><i class="bi bi-trash"></i></div>
                <div class="kit-item"><i class="bi bi-cloud-download" onclick="downloadScreenshoot(${screenshootIdx})"></i></div>
            </div>
        </div>`)
    $('.click-view-pic').click(function () {
        $('#picture-view-modal img').attr('src', $(this).attr('src'))
        $('#picture-view-modal').modal('show')
    })
}

function addVideoTimeNode(action) {
    let id = $("#video-play-id").attr("data-i"),
        time = player.currentTime();
    if (action == 'delete') action = "timedel";
    else action = "timeadd";
    ajaxFunc('/video.json', "POST", { id: id, time: time, action: action }, function (res) {
        showTipToast(res.msg)
        if (res.code == 0) {
            var leftPos = time * 100 / player.duration() + '%';
            if (action == "timeadd") $('.vjs-progress-holder').prepend(`<div class="time-node-item bi bi-geo-alt-fill" onclick="videoPlayJumpTo(${time})" style="left: ${leftPos};"></div>`);
            else $(`.time-node-item[onclick=videoPlayJumpTo(${time})]`).remove();
        }
    })
}

function changeVideoTime(sec) {
    try {
        player.currentTime(player.currentTime() + sec)
    } catch { }
}

function deleteScreenshoot(idx) {
    $(`.screenshoot-list .screenshoot-item[data-i=${idx}]`).remove();
}
function downloadScreenshoot(idx) {
    var formdata = new FormData();
    formdata.append("belong", "screenshot")
    formdata.append("file", screenhootMapFile[screenshootIdx]);
    $.ajax({
        url: "/downloadfile.json"
        , type: "POST"
        , data: formdata
        , processData: false // 告诉jQuery不要去处理发送的数据
        , contentType: false // 告诉jQuery不要去设置Content-Type请求头
        , async: false
        , dataType: "json"
        , success: function (res) {
            showTipToast(res.msg)
            if (res.code == 0) {
                $(`.screenshoot-list .screenshoot-item[data-i=${idx}]`).remove();
            }
        }
    })

}

/****************************************
 *                                      *
 *              MANAGE                  *
 *                                      *
*****************************************/

function categoryAddSubmit() {
    ajaxFunc('/category.json', 'POST', $('#category-form').serialize(), function (res) {
        showTipToast('添加成功')
    })
}

/****************************************
 *                                      *
 *              SPIDER                  *
 *                                      *
*****************************************/

function spiderShtSubmit() {
    var section = $('#spider-sht-form [name=section]').val(),
        day = $('#spider-sht-form [name=day]').val(),
        typeid = $('#spider-sht-form [name=typeid]').val(),
        $resultList = $('#spider-sht tbody'),
        $msglist = $('#spider-sht .spider-msg-list');
    if (section == undefined || section == "" || day == "" || day == undefined) {
        showTipToast("参数不全！")
        return
    }
    var socket = new WebSocket(`ws://${window.location.host}/spider/sht.json?section=${section}&typeid=${typeid}&day=${day}`);
    // Message received on the socket
    socket.onmessage = function (event) {
        var data = JSON.parse(event.data);
        switch (data.type) {
            case 0: // INFO
                switch (data.code) {
                    case 0: // SUCCESS
                        $msglist.append(`<div class="msg-item">${data.msg}</div>`)
                        break;
                    case 1: // ERROR
                        $msglist.append(`<div class="msg-item error">${data.msg}</div>`)
                        break;
                    case 2: // WARN
                        $msglist.append(`<div class="msg-item warn">${data.msg}</div>`)
                        break;
                    case 3: // INFO
                        $msglist.append(`<div class="msg-item info">${data.msg}</div>`)
                        break;
                }
                break;
            case 1: // DATA
                switch (data.code) {
                    case 0: // SUCCESS
                        if ($('#spider-sht tbody tr').length >= 500) {
                            for (let i = 0; i < 450; i++) {
                                $('#spider-sht tbody tr').eq(i).remove();
                            }
                        }
                        $resultList.append(`<tr><td>${data.data.id}</td><td>${data.data.title}</td><td><a target="_blank" href="${data.data.url}">点击链接跳转</a></td><td>${data.data.view}</td><td>${data.data.addtime}</td></tr>`)
                        $('#result-piece-count').html(parseInt($('#result-piece-count').html()) + 1)
                        break;
                    case 1: // ERROR

                        break;
                }
                break;
            case 2: // DONE
                $msglist.append(`<div class="msg-item info">${data.msg}</div>`)
                socket.close()
                break
        }
        $('#spider-sht .spider-result').scrollTop($('#spider-sht .spider-result').prop("scrollHeight"))
        $msglist.scrollTop($msglist.prop("scrollHeight"))
    };
}


function getShtSpiderList(page) {
    ajaxFunc("/spider.json", "GET", {
        mod: 1,
        section: $('#sht-wrap [name=section]').val(),
        category: $('#sht-wrap [name=typeid]').val(),
        page: page,
        filter: $('#sht-wrap [name=filter]').val(),
    }, function (res) {
        if (res.code != 0) {
            showTipToast(res.msg)
        } else {
            let $table = $('#sht-wrap tbody');
            $table.empty();
            for (let obj of res.data)
                $table.append(`<tr><td>${obj.id}</td><td>${obj.title}</td><td><a href="${obj.url}" target="_blank">点击链接跳转</a></td><td>${obj.view}</td><td>${obj.addtime}</td></tr>`);
            if (res.count == undefined) {
                res.count = 0
            }
            setSpiderTablePage('#sht-wrap .table-page', res.count, page)
        }
    });
}

function setSpiderTablePage(idselect, count, page) {
    if (count <= spiderPageThreshold) {
        console.log(count, page)
        $(idselect).empty();
        return ""
    }
    let $pagination = $(idselect),
        pageSize = parseInt(count / spiderLimit),
        pageStart = 1,
        pageThreshold = spiderPageThreshold,
        pageEnd = pageSize;
    $pagination.empty();
    if (pageSize % spiderLimit != 0) pageSize++;
    if (pageSize <= pageThreshold) pageEnd = pageSize;
    else {
        if (page <= pageThreshold / 2) {
            pageEnd = pageStart + pageThreshold
        } else {
            pageStart = page - pageThreshold / 2;
            pageEnd = pageStart + pageThreshold;
            if (pageEnd > pageSize) pageEnd = pageSize;
        }
    }
    $pagination.append(`<a class="page-item page-link" onclick="getShtSpiderList(1)">首页</a>`)
    if (page > 1)
        $pagination.append(`<a class="page-item page-link" onclick="getShtSpiderList(${page - 1})">PRE</a>`);
    for (let i = pageStart; i <= pageEnd; i++) {
        if (i == page) {
            $pagination.append(`<a class="page-item page-link active" href="javascript:;">${i}</a>`)
        } else {
            $pagination.append(`<a class="page-item page-link" onclick="getShtSpiderList(${i})">${i}</a>`);
        }
    }
    if (pageSize > 1 && page < pageSize)
        $pagination.append(`<a class="page-item page-link" onclick="getShtSpiderList(${page + 1})">NEXT</a>`);
    $pagination.append(`<a class="page-item page-link disabled" href="javascript:;">共 ${pageSize} 页</a>`)
}


function openAllLink() {
    var $links = $('#sht-wrap table tbody tr td a');
    for (let i = 0, len = $links.length; i < len; i++) {
        window.open($links.eq(i).attr("href"))
    }
}

/****************************************
 *                                      *
 *              RECORD                  *
 *                                      *
*****************************************/

function getRecord(obj) {
    return `<div class="record-item">
                <input type="checkbox" name="content-show" id="record-${obj.id}">
                <div class="item-kit">
                    <label for="record-${obj.id}" class="kit-item content-toggle bi bi-caret-down"></label>
                    <div class="bi bi-pencil kit-item" onclick="editRecord(${obj.id})"></div>
                </div>
                <div class="record-title">${obj.title}</div>
                <div class="record-detail">${obj.detail}</div>
                <div class="record-content">${obj.content}</div>
            </div>`
}

function setRecordList(res) {
    let $recList = $('#record-page .record-list');
    $recList.empty()
    for (let obj of res.data) {
        $recList.append(getRecord(obj))
    }
}

function recordFormSubmit() {
    let id = $('#record-form [name=id]').val(),
        action = "add";
    if (id != "") {
        action = "edit"
    }
    let data = $('#record-form').serialize();
    data = data + `&action=${action}`
    if ($('#record-form [name=title]').val() == "") {
        showTipToast("所需内容为空！")
        return
    }
    ajaxFunc('/record.json', 'POST', data, function (res) {
        showTipToast(res.msg)
        if (action == "add" && res.code == 0) {
            $('#record-page .record-list').append(getRecord(res.data))
        }
        if (res.code == 0) {
            document.getElementById('record-form').reset();
            $('#record-form').removeClass('was-validated')

        }
    })
}

function editRecord(id) {
    ajaxFunc("/record.json", "GET", { id: id, action: "id" }, function (res) {
        if (res.code == 0) {
            let str = res.data.content;
            str = str.replaceAll('<br/>', '\n')
            res.data.content = str.replaceAll('&ensp;', ' ')
            $('#record-form [name=id]').val(res.data.id)
            $('#record-form [name=title]').val(res.data.title)
            $('#record-form [name=detail]').val(res.data.detail)
            $('#record-form [name=content]').val(res.data.content)
            $(`#record-form [name=category][value=${res.data.category}]`).attr("checked", "checked")
            $('#record-form [name=top]').attr('checked', res.data.top)
            $('#record-page [name=record-show]').prop("checked", true)
        } else {
            showTipToast(res.msg)
        }
    })
}
