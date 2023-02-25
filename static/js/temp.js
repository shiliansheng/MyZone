$(function () {
    


})

var selectTagToInput1 = (id, title) => {
    // if ($('#tag-select-list .tag-item[data-i=' + id + ']').hasClass('tag-select')) {
    //     return
    // }
    console.log($(this).attr('class'))
    if ($(this).hasClass('tag-select')) {
        return
    }
    $('.tag-input-wrap').append(`<div class="tag-item" data-i="${id}"><span>${title}</span><span class="tag-item-close" onclick="excludeTag(${id})">&times;</span></div>`)
    $(this).addClass('tag-select')
}

function selectTagToInput(id, title) {
    let $tagSelect = $('#tag-select-list .tag-item[data-i=' + id + ']')
    if ($tagSelect.hasClass('tag-select')) {
        return
    }
    $('.tag-input-wrap').append(`<div class="tag-item" data-i="${id}"><span>${title}</span><span class="tag-item-close" onclick="excludeTag(${id})">&times;</span></div>`)
    $tagSelect.addClass('tag-select')
}

function excludeTag(id) {
    $('.tag-input-wrap .tag-item[data-i=' + id + ']').remove();
    $('#tag-select-list .tag-item[data-i=' + id + ']').removeClass('tag-select')
}