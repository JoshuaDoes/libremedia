async function loadTranscriptTimings(stream) {
	clearInterval(lyricScrollerId);
	nowPlayingTiming = [];

	if (stream.transcript != null && stream.transcript.lines != null && stream.transcript.lines.length > 0 && nowPlaying != null && nowPlaying.uri == stream.uri && (lastPageUrl == "transcript" || lastPageUrl == "transcript?uri=" + stream.uri)) {
		//console.log("Loading transcript timings for " + stream.uri);

		var lines = stream.transcript.lines;
		if (lines[0].startTimeMs > 0)
			nowPlayingTiming.push([0, ""]);

		for (let i = 0; i < lines.length; i++) {
			var timing = [lines[i].startTimeMs, lines[i].text];
			nowPlayingTiming.push(timing);
		}

		lastScrollY = window.scrollY;
		lastLyric = -1;
		lyricScroller();
		lyricScrollerId = setInterval(lyricScroller, 100);
		//console.log("Spawned auto-scroller: " + lyricScrollerId);
	} else {
		//console.log("Failed to match case for loading transcript timings " + lastPageUrl);
	}
}

function lyricScroll(lyric) {
	//console.log("Scrolling to lyric " + lyric);
	if (nowPlayingTiming.length > 0) {
		var lyricLine;
		if (lyric >= nowPlayingTiming.length) {
			lyricLine = document.getElementById("lyricEnd");
		} else {
			lyricLine = document.getElementById("lyric" + lyric);
		}
		if (lyricLine == null) {
			//console.log("Failed to find lyric line!");
			return;
		}
		//console.log("Scrolling to lyric " + lyric);
		var rect = lyricLine.getBoundingClientRect();
		var absoluteTop = rect.top + window.pageYOffset;
		var middle = absoluteTop - (window.innerHeight / 2);
		window.scrollTo(0, middle + ((rect.bottom - rect.top) / 2));
		lastScrollY = window.scrollY;
		lastLyric = lyric;
	} else {
		//console.log("Now playing timings empty");
		clearInterval(lyricScrollerId);
		nowPlayingTiming = [];
	}
}

function lyricScroller() {
	if (nowPlaying == null) {
		//console.log("Clearing scroller because nothing is playing");
		clearInterval(lyricScrollerId);
		return;
	}

	if (player.paused) {
		//console.log("Auto-scroll is paused");
		return;
	}
	//console.log("Auto-scroll is not paused");

	if (nowPlayingTiming.length > 0) {
		var curTime = player.currentTime*1000;
		var lyric = lastLyric;
		if (lastLyric > -1) {
			for (let i = lastLyric; i < nowPlayingTiming.length; i++) {
				var line = nowPlayingTiming[i];
				if (line == null) {
					continue;
				}
				if (curTime > line[0]) {
					lyric = i;
					continue;
				}
				break;
			}
		} else {
			lyric = 0;
		}
		//console.log("Trying to scroll to lyric " + lyric);
		lyricScroll(lyric);
	}
}

function lyricSeek(lyric) {
	if (nowPlayingTiming.length == 0) {
		return;
	}
	//console.log("Wanting to seek to lyric " + lyric);
	var startTimeMs = 0;
	if (lyric >= nowPlayingTiming.length) {
		startTimeMs = nowPlaying.duration * 1000;
	} else if (lyric >= 0) {
		var line = nowPlayingTiming[lyric];
		startTimeMs = line[0];
	}
	var startTime = Math.floor(startTimeMs/1000);
	//console.log("Seeking to timestamp " + startTime);
	player.currentTime = startTime;
	timer.innerHTML = secondsTimestamp(player.currentTime) + " / " + secondsTimestamp(player.duration);
	lastScrollY = window.scrollY;
	lyricScroll(lyric);
	lyricScrollerId = setInterval(lyricScroller, 100);
	//console.log("Spawned auto-scroller: " + lyricScrollerId);
}