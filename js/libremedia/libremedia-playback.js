function createAudioPlayer() {
	if (createdAudioPlayer) {
		return;
	}
	createdAudioPlayer = true;

	controls.innerHTML = "";
	buttonTranscript = document.createElement("button");
	buttonTranscript.setAttribute("id", "btnTranscript");
	buttonTranscript.innerHTML = '<i class="fa-solid fa-music"></i>';
	buttonPrev = document.createElement("button");
	buttonPrev.setAttribute("id", "btnPrv");
	buttonPrev.innerHTML = '<i class="fa-solid fa-backward-step"></i>';
	buttonPrev.addEventListener("click", playPrev);
	buttonPP = document.createElement("button");
	buttonPP.setAttribute("id", "btnPP");
	buttonPP.innerHTML = play;
	if (!player.paused) {
		buttonPP.innerHTML = pause;
	}
	buttonNext = document.createElement("button");
	buttonNext.setAttribute("id", "btnNxt");
	buttonNext.innerHTML = '<i class="fa-solid fa-forward-step"></i>';
	buttonNext.addEventListener("click", playNext);
	//buttonDownload = document.createElement("button");
	//buttonDownload.setAttribute("id", "btnDownload");
	buttonRepeat = document.createElement("button");
	buttonRepeat.setAttribute("id", "btnRepeat");
	buttonRepeat.innerHTML = '<i class="fa-solid fa-right-long"></i>';
	buttonRepeat.addEventListener("click", toggleRepeat);
	controls.appendChild(buttonTranscript);
	controls.appendChild(buttonPrev);
	controls.appendChild(buttonPP);
	controls.appendChild(buttonNext);
	//controls.appendChild(buttonDownload);
	controls.appendChild(buttonRepeat);

	//console.log("Setting audio events");
	document.getElementById("audioPlayer").addEventListener("ended", playNextEvent);
	audioInit();

	//Rebuild the audio player's metadata if something was playing
	if (nowPlaying != null) {
		const stream = nowPlaying;
		const creator = '<div id="creator"><a href="/creator?uri=' + stream.creators[0].object.uri + '" data-navigo>' + stream.creators[0].object.name + '</a></div>';
		const albumObj = v1GetObject(stream.album.object.uri).object;
		const album = '<div id="album"><a href="/album?uri=' + stream.album.object.uri + '" data-navigo>' + albumObj.name + '</a>';
		//const number = stream.number; //TODO: Return track number of album in the API
		const name = '<div id="stream"><a href="/stream?uri=' + stream.uri + '" data-navigo>' + stream.name + '</a></div>';

		metadata.innerHTML = name + creator + album;
		if (albumObj.datetime != null)
			metadata.innerHTML += '<div id="datetime">(' + albumObj.datetime + ')</div>';
		//buttonDownload.innerHTML = '<a href="/download?uri=' + stream.uri + '" data-navigo><i class="fa-solid fa-download"></i></a>';
		buttonTranscript.innerHTML = '<a href="/transcript" data-navigo><i class="fa-solid fa-music"></i></a>';
		timer.innerHTML = secondsTimestamp(player.currentTime) + " / " + secondsTimestamp(stream.duration);
		navigo.updatePageLinks();
	} else {
		timer.innerHTML = "Waiting to stream...";
		btnPP.innerHTML = loading;
	}
}

function updateAudioPlayer(streamURI) {
	//console.log("Updating audio player " + lastPageUrl);

	if (streamURI == null) {
		player.src = "";
		player.duration = 0;
		resetBgImg();
		return;
	}

	const stream = v1GetObject(streamURI).object;
	nowPlaying = stream;
	const creator = '<div id="creator"><a href="/creator?uri=' + stream.creators[0].object.uri + '" data-navigo>' + stream.creators[0].object.name + '</a></div>';
	const albumObj = v1GetObject(stream.album.object.uri).object;
	const album = '<div id="album"><a href="/album?uri=' + stream.album.object.uri + '" data-navigo>' + albumObj.name + '</a>';
	//const number = stream.number; //TODO: Return track number of album in the API
	const name = '<div id="stream"><a href="/stream?uri=' + streamURI + '" data-navigo>' + stream.name + '</a></div>';
	const duration = stream.duration / 1000.0;

	displayNotification("Now playing:" + name + creator + album, 5000);
	metadata.innerHTML = name + creator + album;
	if (albumObj.datetime != null)
		metadata.innerHTML += '<div id="datetime">(' + albumObj.datetime + ')</div>';
	player.src = "/v1/stream/" + streamURI;
	player.duration = duration;
	//buttonDownload.innerHTML = '<a href="/download?uri=' + streamURI + '" data-navigo><i class="fa-solid fa-download"></i></a>';
	buttonTranscript.innerHTML = '<a href="/transcript" data-navigo><i class="fa-solid fa-music"></i></a>';
	timer.innerHTML = "0:00 / " + secondsTimestamp(duration);
	navigo.updatePageLinks();

	if (lastPageUrl == "transcript" || lastPageUrl == "transcript?uri=" + stream.uri) {
		displayTranscript(null);
	}
	setBgStream(stream);
}

function playStream(match) {
	if (match.params == null) {
		pagePotato(match);
		return;
	}
	var uri = match.params.uri;
	//console.log("Now playing: " + uri);
	updateAudioPlayer(uri);
	queueSet(pageObject);
	audioInit();
	audioPP();
	pagePotato(match);
}

//Replaces the page queue with a new one, skipping forward to nowPlaying and preserving the user's up next queue
function queueSet(pageQueue) {
	//console.log("skipping ahead in queue");
	for (let i = 0; i < pageQueue.length; i++) {
		if (pageQueue[i].uri == nowPlaying.uri) {
			var pageQueueFront = pageQueue.slice(0, i);
			//console.log(pageQueueFront);
			var pageQueueBack = pageQueue.slice(i+1, pageQueue.length);
			//console.log(pageQueueBack);
			pageQueue = pageQueueBack.concat(pageQueueFront);
			//console.log(pageQueue);
			queueLeft = pageQueueBack.length;
			break;
		}
	}
	//console.log(queueLeft);

	if (queueEnd < 0 || queueStart < 0) {
		//console.log("no up next queue, setting page queue as-is");
		queue = pageQueue;
		return;
	}

	//Build the new queue, pushing the up next section to the front
	var newQueue = [];
	if (queueStart == 0) {
		newQueue = queue.slice(0, queueEnd+1);
	} else {
		newQueue = queue.slice(queueStart, queue.length-queueStart);
		if (queueEnd <= queueStart) {
			newQueue = newQueue.concat(queue.slice(0, queueEnd+1));
		}
	}
	queueStart = 0;
	queueEnd = newQueue.length-1;
	queue = newQueue.concat(pageQueue);
	queueLeft += newQueue.length;
}

//Adds to the end of the up next queue
function queueAdd(stream) {
	//console.log("queueAdd-");
	//console.log(queue);
	queueLeft++;
	if (queueEnd < 0 || queueStart < 0) {
		queue.splice(0, 0, stream);
		queueStart = 0;
		queueEnd = 0;
		return;
	}
	queue.splice(queueEnd+1, 0, stream);
	queueEnd++;
	//console.log("queueAdd+");
	//console.log(queue);
}

//Navigo wrapper to add an entry to the queue
function queueAddStream(match) {
	if (match.params == null) {
		pageRelease();
		return;
	}
	var uri = match.params.uri;
	//console.log("Queue: " + uri);
	var stream = {
		"uri": uri,
	}
	queueAdd(stream);
	pagePotato(match);
	displayNotification("Added to stream!", 5000);
}

//Adds to immediately play next, but logically treats it as the end of the up next queue
function queueNext(stream) {
	//console.log("queueNext-");
	//console.log(queue);
	queue.splice(0, 0, stream);
	if (queueEnd == queue.length-1) {
		queueEnd = 0;
	} else {
		queueEnd++;
	}
	//console.log("queueNext+");
	//console.log(queue);
}

//Clears the up next queue
function queueClear() {
	if (queueEnd < 0 || queueStart < 0) {
		return;
	}
	if (queueEnd > queueStart) {
		queue.splice(queueStart, queueEnd+1);
	} else {
		queue.splice(queueStart, queue.length-queueStart-1);
		queue.splice(0, queueEnd+1);
	}
	queueStart = -1;
	queueEnd = -1;
}

//Skips to the next stream in the queue
function playNext() {
	if (nowPlaying == null) {
		return;
	}
	clearInterval(lyricScrollerId);
	lastScrollY = -1;
	lastLyric = -1;
	nowPlayingTiming = [];
	audioPause(); //Pause audio no matter what

	//console.log("playNext-");
	//console.log(queue);

	//If repeating now playing, just restart the stream - user should turn it off to advance
	if (repeat == 2) {
		//updateAudioPlayer(nowPlaying.uri); //TODO: Literally just restart the stream, no need to do all this reloading nonsense but the function already exists and I haven't Googled it yet
		audioPP();
		return;
	}

	//Migrate nowPlaying to end of queue
	queue.push(nowPlaying);

	//If end of queue, we're done!
	if (queueLeft == 0) {
		//console.log("Nothing up next!");
		updateAudioPlayer(null); //TODO: Destroy audio player
		return;
	}
	if (repeat == 0)
		queueLeft--;

	//Migrate next queue entry into nowPlaying
	nowPlaying = queue[0];
	queue.splice(0, 1);

	if (queueEnd > -1 && queueStart > -1) {
		if (queueStart == 0) {
			if (repeat == 1) {
				queueStart = queue.length;
			}
		}
		queueStart--;
		queueEnd--;
	}

	updateAudioPlayer(nowPlaying.uri);
	audioPP();

	//console.log("playNext+");
	//console.log(queue);
}

//Wrapper for playNext
function playNextEvent(event) {
	//console.log("Stream finished! Playing next...");
	playNext();
}

//Returns to the previous stream in the queue, or restarts the song
function playPrev() {
	if (nowPlaying == null) {
		return;
	}
	clearInterval(lyricScrollerId);
	lastScrollY = -1;
	lastLyric = -1;
	nowPlayingTiming = [];
	audioPause(); //Pause audio no matter what

	//console.log("playPrev-");
	//console.log(queue);

	//If repeating now playing, just restart the stream - user should turn it off to advance
	if (repeat == 2) {
		//updateAudioPlayer(nowPlaying.uri); //TODO: Literally just restart the stream, no need to do all this reloading nonsense but the function already exists and I haven't Googled it yet
		audioPP();
		return;
	}

	//Migrate nowPlaying to front of queue
	queue.splice(0, 0, nowPlaying);

	//Migrate last queue entry into nowPlaying
	nowPlaying = queue[queue.length-1];
	queue.splice(queue.length-1, 1);

	if (queueEnd > -1 && queueStart > -1) {
		if (queueStart == queue.length-1) {
			queueStart = -1;
		}
		queueStart++;
		queueEnd++;
	}

	updateAudioPlayer(nowPlaying.uri);
	audioPP();

	//console.log("playPrev+");
	//console.log(queue);
}

//Toggles the current repeat mode
function toggleRepeat() {
	queueLeft = queue.length; //Reset the amount of queue entries left to reflect the remaining queue, even if skipped ahead
	switch (repeat) {
	case 0:
		//console.log("Repeating queue");
		repeat = 1;
		buttonRepeat.innerHTML = '<i class="fa-solid fa-repeat"></i>';
		break;
	case 1:
		//console.log("Repeating now playing");
		repeat = 2;
		buttonRepeat.innerHTML = '<i class="fa-solid fa-hourglass-start"></i>';
		break;
	case 2:
		//console.log("Not repeating");
		repeat = 0;
		buttonRepeat.innerHTML = '<i class="fa-solid fa-right-long"></i>';
		break;
	}
}
