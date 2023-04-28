$(function () {
    $('#video-player-modal').on('hide.bs.modal', function (event) {
        try {
            player.pause()
        } catch { }
        $('#video-player-modal').removeClass('large-modal')
    })
})

function getVideoRecommendList(aim) {
    let date = getFormatDateString(new Date()),
        dataDate = $('#video-recommend-list').attr('data-d');
    if (dataDate == "") {
        dataDate = date;
    }
    if (dataDate == date && aim == "post") {
        return
    }
    date = new Date(dataDate);
    let aimMap = {
        'pre': -1,
        'post': 1,
        'refresh': 0,
    }
    date.setDate(date.getDate() + aimMap[aim])
    date = getFormatDateString(date)
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
                            <img src="${v.cover}" lazyload="on" title="">
                            <span class="video-duration">${v.duration}</span>
                            <span class="video-category">${v.actegorytitle}</span>
                        </a>
                        <div class="video-item-info">
                            <a onclick="playVideoByModal(${v.id})" class="video-title" title="${v.title}">${v.title}</a>
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
        } else {
            showTipToast(res.msg)
        }
        if (aim != "refresh") {
            $('#video-recommend-list').attr('data-d', date);
            $('#video-recommend-list .list-date').html(date);
        }
    })
};

/******* 
 * 改变video modal 大小
 * @return [*] 
 */
function toggleVideoModal() {
    if ($('#video-player-modal').hasClass('large-modal')) {
        $('#video-player-modal').removeClass('large-modal');
    } else {
        $('#video-player-modal').addClass('large-modal');
    }
}

/******* 
 * 配置视频播放控件
 * @param  path [*] 
 * @param  cover [*] 
 * @param  id [*] 
 * @param  timenodes [*] 
 */
var configVideo = function (path, cover, id, timenodes) {
    if (id == playVideoId) {
        return
    }
    if (player != undefined) player.dispose();
    let captionStr = `<track kind="captions" src="${path.replace("mp4", "vtt")}" srclang="zh" label="中文" default>`
    $("#video-player-modal .video-player-container").html(`<video id="video-player" class="video-js vjs-default-skin vjs-matrix vjs-big-play-centered  vjs-16-9 theater-mode" controls preload="auto">${captionStr}</video>`);
    player = videojs('video-player', {
        controls: true,
        loop: false,
        poster: cover.replaceAll('\\', "\\\\"),
        fluid: true,
        liveui: true,
        sources: [{
            src: path,
            type: 'video/mp4'
        }],
        notSupportedMessage: '此视频暂无法播放!', // 无法播放时显示的信息
        playbackRates: [0.5, 1, 2, 4],
        controlBar: {
            progressControl: true,                  // 进度条
            currentTimeDisplay: true,               // 当前时间
            timeDivider: true,                      // 时间分割线
            durationDisplay: true,                  // 总时间
            remainingTimeDisplay: false,            //剩余时间
            customControlSpacer: true,
            playbackRateMenuButton: true,
            fullscreenToggle: true,                 // 全屏按钮
        },
        plugins: {},
    }, function onPlayerReady() {
        let rotateDeg = 0,
            zoom;
        let baseComponent = videojs.getComponent('Component')
        let videoRotateBtn = videojs.extend(baseComponent, {
            constructor: function (thisplayer, options) {
                baseComponent.apply(this, arguments)
                this.on('click', this.clickIcon)
            },
            createEl: function () {
                let divObj = videojs.dom.createEl('div', {
                    className: 'vjs-my-components vjs-control vjs-button',
                    innerHTML: `<button class="bi bi-arrow-repeat vjs-icon-placeholder"></button>`,
                })
                return divObj
            },
            clickIcon: function () {
                rotateDeg += 90;
                zoom = 1;
                if (rotateDeg % 180 != 0) {
                    zoom = document.querySelector('.video-player-container').offsetHeight / document.querySelector('.video-player-container').offsetWidth;
                }
                rotateDeg %= 360;
                player.zoomrotate({
                    rotate: rotateDeg,
                    zoom: zoom,
                });
            }
        })
        let toggleVideo = videojs.extend(baseComponent, {
            constructor: function (thisplayer, options) {
                baseComponent.apply(this, arguments)
                this.on('click', this.clickIcon)
            },
            createEl: function () {
                return videojs.dom.createEl('div', {
                    className: 'vjs-my-components vjs-control vjs-button',
                    innerHTML: `<button class="bi bi-plus-circle vjs-icon-placeholder"></button>`,
                })
            },
            clickIcon: function () {
                toggleVideoModal()
            },
        })

        // 找到 controlBar 节点，添加控件
        videojs.registerComponent('videoRotate', videoRotateBtn)
        player.getChild('controlBar').addChild('videoRotate')
        videojs.registerComponent('toggleVideo', toggleVideo)
        player.getChild('controlBar').addChild('toggleVideo')
        let firstPlay = true;
        this.on('play', function () {//开始播放
            playVideoId = id;
            if (firstPlay) {
                if (timenodes != undefined) {
                    for (let node of timenodes) {
                        let leftPos = node * 100 / this.duration() + '%';
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
};

/******* 
 * video跳转到指定时间
 * @param  time [*] 
 */
function videoPlayJumpTo(time) {
    player.currentTime(time)
};

/******* 
 * 播放视频，指定播放视频ID
 * @param  id [] 
 */
function playVideoByModal(id) {
    $("#video-player-modal #video-relate-list").html('');
    $('.screenshoot-list').html('');
    ajaxFunc("/video.json", "GET", { id: id, action: "videoplay" }, function (res) {
        if (res.code != 0) {
            showTipToast(res.msg)
        } else {
            $("#video-play-id").attr("data-i", res.data.id)
            $("#video-player-modal .modal-title").text(res.data.title)
            configVideo(res.data.path, res.data.cover, res.data.id, res.data.timenodes)
            for (let v of res.data.relateVideos) {
                $("#video-player-modal #video-relate-list").append(
                    `<a onclick="playVideoByModal(${v.id})" class="video-item">
                        <div class="video-cover">
                            <img src="${v.cover}">
                            <div class="video-title">${v.title}</div>
                            <div class="video-duration">${v.duration}</div>
                        </div>
                    </a>`)
            }
            for (let sc of res.data.screenshoots) {
                screenshootIdx++
                $('.screenshoot-list').append(
                    `<div class="screenshoot-item" data-i="${screenshootIdx}">
                        <img class="click-view-pic" src="${sc.path}" title="">
                    </div>`)
                $('.click-view-pic').click(function () {
                    $('#picture-view-modal img').attr('src', $(this).attr('src'))
                    $('#picture-view-modal').modal('show')
                })
            }
        }
    })
    $('#video-player-modal').modal('show');
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

/******* 
 * 设置video edit modal
 * @param  obj [video object] 
 */
function setVideoEidtModal(obj) {
    document.getElementById('videoinfo-form').reset();
    $('#video-edit-modal .video-info-id').text(obj.id);
    $('#video-edit-modal input[name=id]').val(obj.id);
    $('#video-edit-modal input[name=title]').val(obj.title);

    // 设置搜寻标题
    let title = obj.title.toLowerCase();
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
    $(`#tag-select-list-tag .tag-item.tag-select`).removeClass('tag-select');
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


/******* 
 * 上传本地图片作为视频封面，选择文件上传后直接作为视频封面
 */
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

/******* 
 * 爬取视频封面，一般通过选择相应的方式，对视频地址进行爬取，如果使用给定的网址，则会
 * 调用给定的爬取方法，如 javbus.com 等，如果爬取成功，将会将自动添加爬取到的演员和封面
 */
function spiderCover() {
    let title = $('#video-edit-modal #spider-title').val(),
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
            $('#video-edit-modal .video-cover img').attr('src', res.data);
            $(`.video-list .video-item[data-i=${id}] .video-cover img`).attr('src', res.data);
        }
        showTipToast(res.msg)
    })
}

/******* 
 * video information edit form submit
 */
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

/******* 
 * 视频编辑modal显示当前页面视频移动
 * @param  offset [int] 
 */
function videoEditZoom(offset) {
    let $vlist = $('.video-list .video-item'),
        nowIdx = $('#video-edit-modal input[name=id]').val(),
        idx = 0;
    for (let i = 0; i < $vlist.length; i++) {
        if ($vlist.eq(i).attr('data-i') == nowIdx) {
            idx = i + offset
            break
        }
    }
    if (idx < 0 || idx >= $vlist.length) {
        showTipToast('当前是第一个/最后一个！')
        return
    }
    editVideoByModal($vlist.eq(idx).attr('data-i'))
}

/******* 
 * 删除视频，根据视频给定的参数ID进行删除
 * @param  id
 */
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

/******* 
 * 根据 video play modal 中的 #video-play-id 的 data-i 属性获取 video ID
 */
function collectVideo() {
    let id = $("#video-play-id").attr("data-i");
    if (id == "" || id == undefined) {
        return
    }
    ajaxFunc(`/video.json`, "POST", { id: id, action: "collect" }, function (res) {
        showTipToast(res.msg)
    })
}


//////////////// TAG MODULE ////////////////////////

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

/******* 
 * 添加标签
 * @param  obj [string] tag category: tag / actor 
 * @return [*] 
 */
function addTag(obj) {
    let name = "";
    name = $(`#${obj}-add-input`).val();
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

//////////////////////// SCREENSHOT /////////////////////////

/******* 
 * 获取当前视频截图，并添加到视频截图列表中
 */
function shootVideo() {
    const video = document.querySelector('#video-player_html5_api'),
        scale = 1;
    let time = player.currentTime(),
        id = $("#video-play-id").attr("data-i"),
        vname = $('#video-player-modal .modal-title').html();
    let canvas = document.createElement("canvas");
    canvas.width = video.videoWidth * scale, canvas.height = video.videoHeight * scale;
    canvas.getContext('2d').drawImage(video,
        0, 0,
        canvas.width, canvas.height
    );
    let img = document.createElement("img");
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
        `<div class="screenshoot-item" data-i="${screenshootIdx}"><img onclick="viewPicture()" src="${URL.createObjectURL(blob)}" title="">
            <div class="kit-list">
                <div class="kit-item" onclick="deleteScreenshoot(${screenshootIdx})"><i class="bi bi-trash"></i></div>
                <div class="kit-item"><i class="bi bi-cloud-download" onclick="downloadScreenshoot(${screenshootIdx})"></i></div>
            </div>
        </div>`)
}

/******* 
 * 删除截图列表中的对应截图
 * @param  idx [*] 
 */
function deleteScreenshoot(idx) {
    $(`.screenshoot-list .screenshoot-item[data-i=${idx}]`).remove();
}
/******* 
 * 下载视频截图
 * @param  idx [*] 
 */
function downloadScreenshoot(idx) {
    let formdata = new FormData();
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
                $(`.screenshoot-list`).prepend(`<div class="screenshoot-item""><img class="click-view-pic" src="${res.data}" title=""></div>`)
            }
        }
    });
}

/******* 
 * 添加或删除当前时间结点
 */
function addVideoTimeNode(action) {
    let id = $("#video-play-id").attr("data-i"),
        time = player.currentTime();
    action = action == 'delete' ? "timedel" : "timeadd";
    ajaxFunc('/video.json', "POST", { id: id, time: time, action: action }, function (res) {
        showTipToast(res.msg)
        if (res.code == 0) {
            let leftPos = time * 100 / player.duration() + '%';
            if (action == "timeadd") $('.vjs-progress-holder').prepend(`<div class="time-node-item bi bi-geo-alt-fill" onclick="videoPlayJumpTo(${time})" style="left: ${leftPos};"></div>`);
            else $(`.time-node-item[onclick=videoPlayJumpTo(${time})]`).remove();
        }
    })
}

/******* 
 * 跳转视频播放到 sec 时间处
 * @param  sec
 */
function changeVideoTime(sec) {
    try {
        player.currentTime(player.currentTime() + sec)
    } catch { }
}