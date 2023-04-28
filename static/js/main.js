// 定义相关常量
const SUCCESS = 0,
    FAILED = 1,
    STOP = 2;

// 定义相关变量
var module_default = "home",
    category_default = 0,
    videoPageLimit = 18,
    playVideoId,
    player,
    screenshootIdx = 0,
    screenhootMapFile = new Map(),
    // spider list every page limit
    spiderLimit = 10,
    spiderPageThreshold = 10,
    markdownShotObj = {
        htmlDecode: "style,script,iframe",  // you can filter tags decode
        emoji: false,
        taskList: false,
        tex: false,  // 默认不解析
        flowChart: false,  // 默认不解析
        sequenceDiagram: false,  // 默认不解析
        codeFold: true,
    },
    editormdEditObj = {
        width: "100%",
        height: "100%",
        codeFold: true,
        toolbar: true,     //关闭工具栏
        watch: true,       // 关闭实时预览
        path: "/static/lib/editor.md/lib/",
        toolbarIcons: function () {
            return ["undo", "redo", "|", "hr", "del", "ucwords", "uppercase", "lowercase", "|", "preview", "watch", "fullscreen", "|", "image", "table", "datetime", "html-entities", "pagebreak", "search"]
        },
        imageUpload: true,
        imageFormats: ["jpg", "jpeg", "gif", "png", "bmp", "webp"],
        imageUploadURL: "/upload/local?filename=editormd-image-file&belong=mdfile",
    },
    recordEditor;
/******* 
 * ajax common function
 * @param  url [*] 
 * @param  type [string] GET POST
 * @param  data [class] 
 * @param  success [function] 
 */
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

/******* 
 * @param  msg [string] 消息内容
 * @param  time [int] 显示秒数，默认5秒
 * @param  mode [*] 
 */
function showTipToast(msg, time, mode) {
    if (time == undefined) {
        time = 5
    }
    $(`#tip-toast`).attr('data-delay', time * 1000)
    $(`#tip-toast .toast-body`).html(msg)
    $('#tip-toast').toast('show')
}
(function () {
    'use strict';
    // set nav item active
    try {
        let $activeNote = $('#active-note'),
            activeModule = $activeNote.attr('active-module'),
            activeCategory = $activeNote.attr('active-category');
        if (activeModule === "" || activeModule === undefined) activeModule = module_default;
        if (activeCategory === "") activeCategory = category_default;
        $("#top-nav .nav-item.active").removeClass("active")
        $(`#top-nav .nav-item[data-n=${activeModule}]`).addClass("active")
        $("#lside-nav .subnav-item.active").removeClass("active")
        $("#lside-nav .subnav-item[data-i=" + activeCategory + "]").addClass("active")
    } catch { console.log("init module failed...") }

    // 初始化文件上传插件
    try {
        bsCustomFileInput.init()
    } catch { }

    // 设置form验证
    window.addEventListener('load', function () {
        // Fetch all the forms we want to apply custom Bootstrap validation styles to
        let forms = document.getElementsByClassName('needs-validation');
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
        let text = $(this).val().toUpperCase(),
            $tags = $('#tag-select-list-tag .tag-item'),
            len = $tags.length;
        $tags.removeClass('hidden')
        for (let i = 0; i < len; i++) {
            if ($tags.eq(i).html().toUpperCase().indexOf(text) == -1) {
                $tags.eq(i).addClass('hidden')
            }
        }
    })

    $('#actor-add-input').on('input propertychange', function () {
        let text = $(this).val().toUpperCase(),
            $tags = $('#tag-select-list-actor .tag-item'),
            len = $tags.length;
        $tags.removeClass('hidden')
        for (let i = 0; i < len; i++) {
            if ($tags.eq(i).html().toUpperCase().indexOf(text) == -1) {
                $tags.eq(i).addClass('hidden')
            }
        }
    })

    $('#actor-filter-input').on('input propertychange', function () {
        let text = $(this).val().toUpperCase(),
            $tags = $('#side-list-actor .tag-item'),
            $tagNames = $('#side-list-actor .tag-item a'),
            len = $tags.length;
        $tags.removeClass('hidden')
        for (let i = 0; i < len; i++) {
            if ($tagNames.eq(i).html().toUpperCase().indexOf(text) == -1) {
                $tags.eq(i).addClass('hidden')
            }
        }
    })

    // 跳转到最上面和最下面
    $('.to-down-kit').on('click', function () {
        let $elem = $($(this).attr('data-target') + '');
        $elem.scrollTop($elem.prop("scrollHeight"))
    })
    $('.to-top-kit').on('click', function () {
        let $elem = $($(this).attr('data-target') + '');
        $elem.scrollTop(-$elem.prop("scrollHeight"))
    })
    // 动态选择select子父级
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
    // 图片点击展示
    $('.click-view-pic').click(function () {
        console.log($(this).attr('src'))
        $('#picture-view-modal img').attr('src', $(this).attr('src'))
        $('#picture-view-modal').modal('show')
    })
})();


/******* 
 * 返回2006-01-02格式日期 
 * @param  date [Date] 
 * @return [string] 
 */
function getFormatDateString(date) {
    return date.getFullYear() + '-' + get2LenNum((date.getMonth() + 1) + '') + '-' + get2LenNum(date.getDate() + '')
}

/******* 
 * 给定数字返回至少2位字符串，少则前补0
 * @param  num [*] 
 * @return [string] 
 */
function get2LenNum(num) {
    if (num.length == 1) {
        return '0' + num;
    }
    return num;
}


/******* 
 * 图片点击获取当前图片的src并展示在 picture-view-modal 中
 */
function viewPicture() {
    $('#picture-view-modal img').attr('src', $(this).attr('src'))
    $('#picture-view-modal').modal('show')
}


/******* 
 * 给指定ID的元素进行 toggle 显示大小
 * 主要方式为 添加 / 删除 .full-screen
 * @param  id [*] 
 */
function toggleScreen(id) {
    let fullScreenClass = 'full-screen';
    if ($(`${id}`).hasClass(fullScreenClass)) {
        $(`${id}`).removeClass(fullScreenClass);
    } else {
        $(`${id}`).addClass(fullScreenClass);
    }
}