function setBgImg(url) {
	if (url == "")
		return;

	//console.log("setBgImg: " + url);
	var newbg = 'url("' + url + '")';
	if (url == "") {
		newbg = "";
	}

	if (document.body.style.backgroundImage !== newbg) {
		console.log("Setting background " + url + " using " + newbg + " to replace " + document.body.style.backgroundImage);
		document.body.style.backgroundImage = newbg;
		bgImg = url;
	}
}

function setBgStream(stream) {
	if (stream.album.object.artworks != null && stream.album.object.artworks.length > 0) {
		var selbg = stream.album.object.artworks.length - 1;
		if (selbg > 4) {
			selbg = 4;
		}
		const bestbg = stream.album.object.artworks[selbg];
		setBgImg(bestbg.url);
	}
}

function resetBgImg() {
	//console.log("resetBgImg: " + bgImg);
	var newbg = '';
	if (nowPlaying != null) {
		if (nowPlaying.album.object.artworks != null) {
			var selbg = nowPlaying.album.object.artworks.length - 1;
			if (selbg > 4) {
				selbg = 4;
			}
			const bestbg = nowPlaying.album.object.artworks[selbg];
			newbg = bestbg.url;
		}
	}
	if (newbg == "") {
		newbg = bgImg;
	}
	setBgImg(newbg);
}