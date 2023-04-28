(function () {
    // 设置侧边导航栏
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

    //初始化编辑form
    recordEditor = editormd('record-editor', editormdEditObj)

    // 初始化各个record item，将md转换为html
    var $recordlist = $('.record-list .record-item .record-content');
    for (let i = 0; i < $recordlist.length; ++i) {
        let id = $recordlist.eq(i).attr('id');
        editormd.markdownToHTML(id, markdownShotObj);
    }
})();

function getRecord(obj) {
    let str = `<div class="record-item" data-i="${obj.id}">
                <input type="checkbox" name="content-show" id="record-${obj.id}">
                <div class="item-kit">
                    <label for="record-${obj.id}" class="kit-item content-toggle bi bi-caret-down-square"></label>
                    <div class="bi bi-pencil-fill kit-item" onclick="editRecord(${obj.id})"></div>
                    <div class="bi bi-clipboard-check-fill kit-item" onclick="copyRecord(${obj.id})"></div>
                </div>
                <div class="record-title">${obj.title}</div>
                <div class="record-detail">${obj.detail}</div>
                <div class="record-content" id="record-content-${obj.id}"><textarea style="display: none;">${obj.content}</textarea></div>
            </div>`
    return str;
}

function setRecordList(res) {
    let $recList = $('#record-page .record-list');
    $recList.empty()
    for (let obj of res.data) {
        $recList.append(getRecord(obj))
        editormd.markdownToHTML(`record-content-${obj.id}`, markdownShotObj);
    }
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
            $('#record-form [name=content]').html(res.data.content)
            $('#record-form [name=content]').val(res.data.content)
            recordEditor = editormd('record-editor', editormdEditObj)
            $(`#record-form [name=category][value=${res.data.category}]`).attr("checked", "checked")
            $('#record-form [name=top]').attr('checked', res.data.top)
            $('#record-page [name=record-show]').prop("checked", true)
        } else {
            showTipToast(res.msg)
        }
    })
}

// copy content
// selector: .record-list .record-item[data-i=${id}] .record-content
function copyRecord(id) {
    // 选择指定元素并获取其文本内容
    let content = $(`.record-list .record-item[data-i=${id}] .record-content`).text();
    // 创建一个临时的textarea元素并将文本内容赋值给它
    let $temp = $('<textarea>');
    $('body').append($temp);
    $temp.val(content).select();
    // 调用浏览器的复制命令将文本复制到剪贴板中
    document.execCommand('copy');
    // 将临时元素删除
    $temp.remove();
    showTipToast('成功拷贝到剪贴板中！');
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
            editormd.markdownToHTML(`record-content-${res.data.id}`, markdownShotObj);
        }
        if (action == "edit" && res.code == 0) {
            $(`#record-content-${id}`).html(`<textarea style="display: none;">${res.data.content}</textarea>`)
            editormd.markdownToHTML(`record-content-${id}`, markdownShotObj);
        }
        if (res.code == 0) {
            recordEditor.clear();
            document.getElementById('record-form').reset();
            $('#record-form').removeClass('was-validated')
        }
    })
}