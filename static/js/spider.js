
/******* 
 * 获取 色花堂 spider list
 * @param  page [*] 
 */
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
            setSpiderTablePage('#sht-wrap .table-page', res.count, page, 'Sht')
        }
    });
}

/******* 
 * 首页获取 2048 spider list
 * @param  page [*] 
 */
function get2048SpiderList(page) {
    ajaxFunc("/spider.json", "GET", {
        mod: 3,
        section: 0,
        category: $('#2048-wrap [name=typeid]').val(),
        page: page,
        filter: $('#2048-wrap [name=filter]').val(),
    }, function (res) {
        if (res.code != 0) {
            showTipToast(res.msg)
        } else {
            let $table = $('#2048-wrap tbody');
            $table.empty();
            for (let obj of res.data)
                $table.append(`<tr><td>${obj.id}</td><td>${obj.title}</td><td><a href="${obj.url}" target="_blank">点击链接跳转</a></td><td>${obj.pubtime}</td><td>${obj.addtime}</td></tr>`);
            if (res.count == undefined) {
                res.count = 0
            }
            setSpiderTablePage('#2048-wrap .table-page', res.count, page, '2048')
        }
    });
}

function setSpiderTablePage(idselect, count, page, aim) {
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
    if ((pageSize % spiderLimit) != 0) pageSize++;
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
    $pagination.append(`<a class="page-item page-link" onclick="get${aim}SpiderList(1)">首页</a>`)
    if (page > 1)
        $pagination.append(`<a class="page-item page-link" onclick="get${aim}SpiderList(${page - 1})">PRE</a>`);
    for (let i = pageStart; i <= pageEnd; i++) {
        if (i == page) {
            $pagination.append(`<a class="page-item page-link active" href="javascript:;">${i}</a>`)
        } else {
            $pagination.append(`<a class="page-item page-link" onclick="get${aim}SpiderList(${i})">${i}</a>`);
        }
    }
    if (pageSize > 1 && page < pageSize)
        $pagination.append(`<a class="page-item page-link" onclick="get${aim}SpiderList(${parseInt(page) + 1})">NEXT</a>`);
    $pagination.append(`<a class="page-item page-link disabled jump-item" onclick="jumpToSpiderPage('${idselect}')" data-aim="${aim}" data-maxpage="${pageSize}"><input class="page-min-input" name="page">/${pageSize}</a>`)
}

/******* 
 * 跳转到输入的页数
 * @param  pid [*] 相应的pageid 如 sht / 2048
 */
function jumpToSpiderPage(pid) {
    let aim = $(`${pid} .jump-item`).attr('data-aim'),
        maxpage = $(`${pid} .jump-item`).attr('data-maxpage'),
        page = $(`${pid} .jump-item input[name=page]`).val();
    if (page == "" || parseInt(page) > maxpage || parseInt(page) <= 0) {
        return
    }
    if (aim == "Sht") {
        getShtSpiderList(page)
    } else if (aim == "2048") {
        get2048SpiderList(page)
    }
}

/******* 
 * 打开aim(sht/2048) spider list 当前页中的所有链接到新标签页中
 * @param  aim [*] 
 */
function openAllLink(aim) {
    let $links = $(`#${aim}-wrap table tbody tr td a`);
    for (let i = 0, len = $links.length; i < len; i++) {
        window.open($links.eq(i).attr("href"))
    }
}

//////////////////////////////// SPIDER EVERY ITEM ///////////////////////////////////


/******* 
 * 爬取色花堂
 * 具体流程包括：
 *      获取需要爬取的相关信息
 *      连接websocket，进行爬取
 *      获取信息，根据信息中相应的type选择不同的操作方式
 *      如果type==INFO，将信息显示在 msg show panel 中
 *      如果type==DATA，则进行相关的数据处理
 */
function spiderShtSubmit() {
    let section = $('#spider-sht-form [name=section]').val(),
        day = $('#spider-sht-form [name=day]').val(),
        typeid = $('#spider-sht-form [name=typeid]').val(),
        $resultList = $('#spider-sht tbody'),
        $msglist = $('#spider-sht .spider-msg-list');
    if (section == undefined || section == "" || day == "" || day == undefined) {
        showTipToast("参数不全！")
        return
    }
    let socket = new WebSocket(`ws://${window.location.host}/spider/sht.json?section=${section}&typeid=${typeid}&day=${day}`);
    // Message received on the socket
    socket.onmessage = function (event) {
        let data = JSON.parse(event.data);
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



function spiderUaaSubmit() {
    let nids = $('#spider-uaa [name=nids]').val(),
        $resultList = $('#spider-uaa .result-view tbody'),
        $msglist = $('#spider-uaa .spider-msg-list'),
        $progress = $('#spider-progress .progress-bar');
    if (nids == "" || nids == undefined) {
        showTipToast("参数不全！")
        return
    }
    nids = nids.replaceAll('\n', ' ')
    const socket = new WebSocket(`ws://${window.location.host}/spider/uaa.json?nids=${nids}`);
    socket.onopen = function (event) {
        console.log('WebSocket 连接已建立');
    };
    // Message received on the socket
    socket.onmessage = function (event) {
        let data = JSON.parse(event.data);
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
                        if ($('#spider-uaa tbody tr').length >= 500) {
                            for (let i = 0; i < 450; i++) {
                                $('#spider-uaa tbody tr').eq(i).remove();
                            }
                        }
                        $resultList.append(`<tr><td>${data.data.id}</td><td>${data.data.title}</td><td>${data.data.content}</td></tr>`)
                        $('#result-piece-count').html(parseInt($('#result-piece-count').html()) + 1)
                        $progress.attr('aria-valuenow', parseInt($progress.attr('aria-valuenow')) + 1)
                        let progressNum = parseFloat($progress.attr('aria-valuenow')) / parseFloat($progress.attr('aria-valuemax')) * 100;
                        progressNum = progressNum.toFixed(2) + '%'
                        $progress.css('width', progressNum)
                        $progress.html(progressNum)

                        break;
                    case 1: // ERROR
                        break;
                    case 3: // INFO
                        $('#spider-uaa .data-overview tbody').prepend(`<tr><td>${data.data.id}</td><td>${data.data.title}</td><td>${data.data.authors}</td><td>${data.data.spiderTime}</td></tr>`)
                        $('#result-piece-count').html(0)
                        $('#result-piece-title').html(data.data.title)
                        $('#result-piece-chaptercount').html(data.data.chapterCount)
                        $progress.attr('aria-valuemax', data.data.chapterCount)
                        $progress.attr('aria-valuenow', 0)

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
    socket.onclose = function (event) {
        console.log('WebSocket 连接已关闭');
    };
    socket.onerror = function (event) {
        console.error('WebSocket 连接错误');
    };
}


function spider2048Submit() {
    let day = $('#spider-2048-form [name=day]').val(),
        typeid = $('#spider-2048-form [name=typeid]').val(),
        $resultList = $('#spider-2048 tbody'),
        $msglist = $('#spider-2048 .spider-msg-list');
    if (day == "" || day == undefined) {
        showTipToast("参数不全！")
        return
    }
    const socket = new WebSocket(`ws://${window.location.host}/spider/2048.json?typeid=${typeid}&day=${day}`);
    socket.onopen = function (event) {
        console.log('WebSocket 连接已建立');
    };
    // Message received on the socket
    socket.onmessage = function (event) {
        let data = JSON.parse(event.data);
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
                        if ($('#spider-2048 tbody tr').length >= 500) {
                            $('#spider-2048 tbody').empty()
                        }
                        $resultList.append(`<tr><td>${data.data.id}</td><td>${data.data.title}</td><td><a target="_blank" href="${data.data.url}">点击链接跳转</a></td><td>${data.data.pubtime}</td><td>${data.data.addtime}</td></tr>`)
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
        $('#spider-2048 .spider-result').scrollTop($('#spider-2048 .spider-result').prop("scrollHeight"))
        $msglist.scrollTop($msglist.prop("scrollHeight"))
    };
    socket.onclose = function (event) {
        console.log('WebSocket 连接已关闭');
    };
    socket.onerror = function (event) {
        console.error('WebSocket 连接错误');
    };
}