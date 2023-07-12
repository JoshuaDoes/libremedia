function clearPage() {
	//console.log("clearPage");
	clearTimeout(delayTimer);
	delayTimer = null;
	clearInterval(lyricScrollerId); //User has definitely navigated away from transcript page
	lyricScrollerId = null;
	resetScroll();
	lastScrollY = -1;
	lastLyric = -1;
	render(null, ""); //Clear the page
	if (searching != null)
		searching.innerHTML = '';
	if (searchbox != null)
		searchbox.value = '';
	bgImg = "";
	resetBgImg();
	//pageObject = [];
}

async function displayNotification(msg, timeout) {
	//console.log("Notification: " + msg);
	interruptNotification = true;
	if (notif == null)
		return;
	notif.innerHTML = "";
	notif.style.opacity = "0.0";
	await new Promise(r => setTimeout(r, 33));
	interruptNotification = false;
	notif.innerHTML = msg;
	for (let i = 1; i < 10; i++) {
		if (interruptNotification)
			return;
		notif.style.opacity = "0." + i;
		await new Promise(r => setTimeout(r, 33));
	}
	if (interruptNotification)
		return;
	notif.style.opacity = "1";
	await new Promise(r => setTimeout(r, timeout));
	for (let i = 9; i >= 0; i--) {
		if (interruptNotification)
			return;
		notif.style.opacity = "0." + i;
		await new Promise(r => setTimeout(r, 33));
	}
	if (interruptNotification)
		return;
	notif.innerHTML = "";
}

async function displaySearch(match) {
	var q = match.params.q;
	//console.log("Search: " + q);
	document.querySelector("#searching").innerHTML = '🔎';
	clearTimeout(delayTimer);
	delayTimer = setTimeout(function() {
		var results = v1GetSearch(q).object;
		pageObject = results.streams;

		var html = "";
		//Streams
		if (results.streams != null)
			html += tblStreams(results.streams);

		//Creators
		if (results.creators != null)
			html += tblCreators(results.creators);

		//Albums
		if (results.albums != null)
			html += tblAlbums(results.albums);

		if (delayTimer == null)
			return;
		render(match, html);
		searchbox.focus();
		searchbox.value = q;
	}, 1000);
}

async function displayCreator(match) {
	if (match.params == null) {
		pageRelease();
		return;
	}
	var uri = match.params.uri;
	//console.log("Creator: " + uri);

	var creator = v1GetObject(uri);

	if (creator.object.artworks != null && creator.object.artworks.length > 0) {
		const bestbg = creator.object.artworks[creator.object.artworks.length-1];
		setBgImg(bestbg.url);
	}

	var html = '<tr><th colspan="5"><div id="creator"><a href="/creator?uri=' + creator.uri + '" data-navigo>' + creator.object.name + '</a></div></th></tr>';
	if (creator.object.genres != null && creator.object.genres.length > 0) {
		html += '<tr><td><div id="genre">';
		for (let i = 0; i < creator.object.genres.length; i++) {
			html += creator.object.genres[i] + '<br />';
		}
		html += '</div></td></tr>';
	}
	html += '</tr>';
	if (creator.object.description != null && creator.object.description.length > 0) {
		const bio = creator.object.description
			.replace(/\r\n|\r|\n/gim, '<br>') // linebreaks
			.replace(/\[([^\[]+)\](\(([^)]*))\)/gim, '<a href="$3" data-navigo>$1</a>'); // anchor tags
		const splitBio = bio.split(" ");
		html += '<tr><td colspan="5">';
		if (splitBio.length > 20) {
			smallBioView = splitBio.slice(0, 20).join(" ");
			smallBioHidden = smallBioView + " " + splitBio.slice(20, splitBio.length).join(" ");
			html += '<p id="readMore" onclick="textExpander()">' + smallBioView + ' ...<br /><small>(tap to show more)</small></p><span id="more"><p id="showLess" onclick="textExpander()">' + smallBioHidden + '<br /><small>(tap again to show less)</small></p></span>';
		} else {
			html += '<div>' + bio + '</div>';
		}
		html += '</td></tr>';
	}

	//Top tracks
	if (creator.object.topStreams != null && creator.object.topStreams.length > 0) {
		pageObject = creator.object.topStreams;
		html += tblStreamsTop(creator.object.topStreams);
	}

	//Albums
	if (creator.object.albums != null && creator.object.albums.length > 0) {
		html += tblAlbums(creator.object.albums);
	}

	//Appearances
	if (creator.object.appearances != null && creator.object.appearances.length > 0) {
		html += tblAppearances(creator.object.appearances);
	}

	//Singles
	if (creator.object.singles != null && creator.object.singles.length > 0) {
		html += tblSingles(creator.object.singles);
	}

	//Related
	if (creator.object.related != null && creator.object.related.length > 0) {
		html += tblRelated(creator.object.related);
	}

	render(match, html);
}

async function displayAlbum(match) {
	if (match.params == null) {
		pageRelease();
		return;
	}
	var uri = match.params.uri;
	//console.log("Album: " + uri);

	var album = v1GetObject(uri);

	if (album.object.artworks != null && album.object.artworks.length > 0) {
		var selbg = album.object.artworks.length-1;
		if (selbg > 4) {
			selbg = 4;
		}
		const bestbg = album.object.artworks[selbg];
		setBgImg(bestbg.url);
	}

	var html = '';
	if (album.object.creators != null) {
		html += '<tr><th colspan="2"><div id="creator"><a href="/creator?uri=' + album.object.creators[0].uri + '" data-navigo>' + album.object.creators[0].object.name + '</a></div></th></tr>';
	}
	html += '<tr><th colspan="2"><div id="album"><a href="/album?uri=' + album.uri + '" data-navigo>' + album.object.name;
	if (album.object.datetime != null) {
		html += '<br /><div id="datetime">(' + album.object.datetime + ')</div>';
	}
	html += '</a></div></th></tr>';

	//Discs
	for (let i = 0; i < album.object.discs.length; i++) {
		html += '<tr><th colspan="2">Disc ' + (i + 1) + '</th></tr><tr><th>Streams (' + album.object.discs[i].streams.length + ')</th><th>🕑</th></tr>';
		for (let j = 0; j < album.object.discs[i].streams.length; j++) {
			pageObject.push(album.object.discs[i].streams[j].object);
			html += tblStreamAlbum(album.provider, album.object.discs[i].streams[j].object, j+1);
		}
		html += '<br />';
	}

	render(match, html);
}

async function displayTranscript(match) {
	clearInterval(lyricScrollerId);
	nowPlayingTiming = [];

	var uri = "";
	if (nowPlaying != null) {
		uri = nowPlaying.uri;
	}
	if (match != null && match.params != null) {
		uri = match.params.uri;
	}
	if (uri == "") {
		pageRelease();
		return;
	}
	//console.log("Transcript: " + uri);

	var stream = nowPlaying;
	var isNowPlaying = true;
	if (stream == null || uri != stream.uri) {
		isNowPlaying = false;
		stream = v1GetObject(uri).object;
	}
	//console.log("Transcript: Is now playing? " + isNowPlaying);

	if (stream.transcript != null && stream.transcript.lines != null && stream.transcript.lines.length > 0) {
		var colspan = 0;
		var html = '';
		if (stream.creators != null) {
			html += '<tr><th><div id="creator"><a href="/creator?uri=' + stream.creators[0].uri + '" data-navigo>' + stream.creators[0].object.name + '</a></div></th>';
			colspan++;
		}
		if (stream.album != null) {
			if (colspan == 0) {
				html += '<tr>';
			}
			html += '<th><div id="album"><a href="/album?uri=' + stream.album.uri + '" data-navigo>' + stream.album.object.name + '</a></div></th>';
			colspan++;
		}
		if (colspan > 0) {
			html += '</tr>';
		}
		colspan++;
		html += '<tr><th colspan="' + colspan + '"><div id="stream"><a href="/stream?uri=' + uri + '" data-navigo>' + stream.name + '</a></div></th></tr>';

		var lines = stream.transcript.lines;
		//console.log("Building transcript for " + lines.length + " lines, should add timings? " + isNowPlaying);
		html += '<tr><td id="lyricblank" onclick="lyricSeek(0)" colspan="' + colspan + '"><div id="lyric0"></div></td></tr>';
		for (let i = 0; i < lines.length; i++) {
			if (lines[i].text != null) {
				html += '<tr><td id="lyric" onclick="lyricSeek(' + (i+1) + ')" colspan="' + colspan + '"><div id="lyric' + (i+1) + '">' + lines[i].text + '</div></td></tr>';
			} else {
				html += '<tr><td id="lyricblank" onclick="lyricSeek(' + (i+1) + ')" colspan="' + colspan + '"><div id="lyric' + (i+1) + '"></div></td></tr>';
			}
		}
		html += '<tr><td id="lyricblank" colspan="' + colspan + '" onclick="lyricSeek(' + lines.length + ')"><div id="lyricEnd"></div></td></tr>';
		render(match, html);

		if (isNowPlaying) {
			//console.log("Now playing, loading transcript timings");
			loadTranscriptTimings(stream);
		}
	} else {
		render(match, "No transcript for this stream!");
	}
}
