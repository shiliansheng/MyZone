function categoryAddSubmit() {
    ajaxFunc('/category.json', 'POST', $('#category-form').serialize(), function (res) {
        showTipToast('添加成功')
    })
}



///////////////////// TAG //////////////////////////

function editTag(type, id) {
    ajaxFunc(`/${type}.json`, "GET", { id: id, obj: type }, function (res) {
        if (res.code == 0) {
            $('#tag-edit-form').attr('data-type', type);
            $('#tag-edit-form [name=id]').val(res.data.id);
            $('#tag-edit-form [name=name]').val(res.data.name);
        }
        showTipToast(res.msg)
    })
}

function delTag(type, id) {
    if (confirm(`是否删除该标签[${type} - ${id}]？`)) {
        ajaxFunc(`/${type}.json?id=${id}&obj=${type}`, "DELETE", {}, function (res) {
            showTipToast(res.msg)
            if (res.code == 0) {
                $(`.tag-item[data-i=${type}-${id}]`).remove();
            }
        })
    }
}


function tagEditSubmit() {
    let type = $('#tag-edit-form').attr('data-type'),
        id = $('#tag-edit-form [name=id]').val();
    ajaxFunc(`/${type}.json`, "POST", {
        id: id,
        name: $('#tag-edit-form [name=name]').val(),
        action: "edit",
        obj: type,
    }, function (res) {
        if (res.code == 0) {
            $(`.tag-item[data-i=${type}-${id}] a`).html(res.data)
        }
        showTipToast(res.msg)
    })
}