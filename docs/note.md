# NOTE

## 一、GO

### 执行带空格的CMD命令
```go
cmd := exec.Command("cmd")
cmd.SysProcAttr = &syscall.SysProcAttr{CmdLine: fmt.Sprintf(`/c %s`, cmdString), HideWindow: true}
out, err := cmd.CombindOutput()
```
## 二、CSS

### transition

对于隐藏元素时，和`transition`进行搭配使用，对于隐藏元素的方法则有限制  
隐藏元素的方法：

-   `display: none`: 这种方法会让元素直接从渲染树中消失，无法使用`transition`渲染
-   `visibility:hidden/visible`: 这种方法不会让元素从渲染树中消失，使用`transition`时只能渲染`hidden`，无点击等事件
-   `opacity: 0/1`: 这种方法可以渲染，但是元素仍在原处，有点击等事件
-   `position:absolute left:-9999px或top:-9999px;`
-   通过设置`width:0+font-size:0;`或`height:0+font-size:0`
-   `z-index`为负值

### 元素四角边框

```css
background: linear-gradient(to left, var(--blue), var(--blue)) left top no-repeat,
	linear-gradient(to bottom, var(--blue), var(--blue)) left top no-repeat,
	linear-gradient(to left, var(--blue), var(--blue)) right top no-repeat, linear-gradient(
			to bottom,
			var(--blue),
			var(--blue)
		) right top no-repeat,
	linear-gradient(to left, var(--blue), var(--blue)) left bottom no-repeat, linear-gradient(
			to bottom,
			var(--blue),
			var(--blue)
		) left bottom no-repeat,
	linear-gradient(to left, var(--blue), var(--blue)) right bottom no-repeat, linear-gradient(
			to left,
			var(--blue),
			var(--blue)
		) right bottom no-repeat;
background-size: 3px 10px, 10px 3px, 3px 10px, 10px 3px;
```

### 玻璃拟态

分为三步：模糊(`backdrop-filter: blur(npx);`、投影(`box-shawdow`)、加点白(`background: rgba(255 255 255 / .3)`)

```css
backdrop-filter: blur(5px);
box-shadow: 0 8px 12px rgba(255, 255, 255, 0.3);
background: rgba(255, 255, 255, 0.5);
```

### 文本省略

单行省略
```css
.title {
	width: 100px;
	text-overflow: ellipsis;
	overflow: hidden;
	white-space: nowrap;
}
```
多行省略
```css
.title {
	width: 100px;
    word-break: break-all;
    text-overflow: ellipsis;
    overflow: hidden;
    display: -webkit-box;
    -webkit-box-orient: vertical;
    -webkit-line-clamp: 2; /* 这里是超出几行省略 */
}
```

## 三、JS

### AJAX 加载 HTML 页面

加载一个 html 的话是可以分为加载其中某个块(div)和加载整个页面，而不管加载其中任何一种都是需要本页面的一个块(div)来进行加载展示。加载的方法可以是 `$(ajax{})` 方法也可以是 `$('#div').load()` 方法  
1、加载整个页面

```js
$.ajax({
	url: "./test.html",
	type: "get",
	success: function (res) {
		$("#router").html($(res));
	},
});
```

2、加载部分内容

```js
$.ajax({
	url: "./test.html",
	type: "get",
	success: function (res) {
		var html = $(res).find(".warp");
		$("#router").html(html);
	},
});
```

### AJAX跨域

### 获取 video 任意时间截图

```html
<div contenteditable="true" id="in-box"></div>
<div>
	<input type="file" name="" id="upload-ipt" />
	<div class="review" id="out-box"></div>
</div>
```

```js
function getVideoImage() {
	var obj_file = document.getElementById("upload-ipt");
	var file = obj_file.files[0];
	var blob = new Blob([file]), // 文件转化成二进制文件
		url = URL.createObjectURL(blob); //转化成url
	if (file && /video/g.test(file.type)) {
		var $video = $(
			'<div><video controls src="' +
				url +
				'"></video></div><div>&nbsp;</div>'
		);
		//后面加一个空格div是为了解决在富文本中按Backspace时删除无反应的问题
		$("#in-box").html($video);
		var videoElement = $("video")[0];
		videoElement.addEventListener("canplay", function (_event) {
			var canvas = document.createElement("canvas");
			canvas.width = videoElement.videoWidth;
			canvas.height = videoElement.videoHeight;
			console.log(videoElement.videoWidth);
			canvas
				.getContext("2d")
				.drawImage(videoElement, 0, 0, canvas.width, canvas.height);
			var img = document.createElement("img");
			img.src = canvas.toDataURL("image/png");
			$("#out-box").html(img);
			URL.revokeObjectURL(this.src); // 释放createObjectURL创建的对象
			console.log("loadedmetadata");
		});
	} else {
		alert("请上传一个视频文件！");
		obj_file.value = "";
	}
}
```

### base64 转换为 blob

```js
//ndata为base64格式地址
let arr = ndata.split(","),
	mime = arr[0].match(/:(.*?);/)[1],
	bstr = atob(arr[1]),
	n = bstr.length,
	u8arr = new Uint8Array(n);
while (n--) {
	u8arr[n] = bstr.charCodeAt(n);
}
let bdata = new Blob([u8arr], { type: mime });
```

### 设置checkbox
```js
$('#checkboxId').prop('checked', true)
$('#checkboxId').prop('checked', false)
if ($('#checkedboxId').is('checked')) {
	// do something
}
```

## 四、MYSQL

### 查找排序顺序错乱问题

问题描述：mysql 对⽆索引字段进⾏排序后 limit ，当被排序字段有相同值时并且在 limit 范围内，取的值并不是正常排序后的值，有可能第⼀页查询的记录，重复出现在第⼆页的查询记录中，⽽且第⼆页的查询结果乱序，导致分页结果查询错乱问题。  
问题解决：`order by` 后多添加⼀个`id`字段排序
SQL 语句为：

```sql
SELECT id,word,nature,weight,order_num FROM unlp_hot_dictionary ORDER BY order_num, id DESC LIMIT 0,10;
SELECT id,word,nature,weight,order_num FROM unlp_hot_dictionary ORDER BY order_num, id DESC LIMIT 10,10;
```

## 五、HTML

### Chrome 图片懒加载

```html
<img src="#" lazyload="on">
```