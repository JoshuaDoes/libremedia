//Navigation map
var navMap = {
	"search": displaySearch,
	"creator": displayCreator,
	"album": displayAlbum,
	"transcript": displayTranscript,
	"addqueue": queueAddStream,
	"stream": playStream,
	"download": downloadStream,
}

function navigoResolve() {
	navigo = new Navigo("/", { hash: true });
	navigo
		.on("/search", (match) => {
			pageObject = [];
			navMap["search"](match);
		})
		.on("/creator", (match) => {
			pageObject = [];
			navMap["creator"](match);
		})
		.on("/album", (match) => {
			pageObject = [];
			navMap["album"](match);
		})
		.on("/transcript", (match) => {
			navMap["transcript"](match);
		})
		.on("/addqueue", (match) => {
			navMap["addqueue"](match);
		})
		.on("/stream", (match) => {
			navMap["stream"](match);
		})
		.on("/download", (match) => {
			navMap["download"](match);
		})
		.on("/back", (match) => {
			pageRelease();
		})
		.notFound((match) => {
			render(match, '<center>404<br />' + match.url + '<br />' + match.hashString + '</center><br /><a href="/back" data-navigo>Return to where you came from</a> or search for something to stream!');
		})
		.on((match) => {
			//console.log("Nothing to do!");
			render(match, "");
		})
	.resolve();
}

function refreshElements() {
	content = document.getElementById("content");
	infobar = document.getElementById("infobar");
	if (infobar === null) {
		infobar = document.getElementById("infobar hidden");
	}
	timer = document.getElementById("audioTimer");
	player = document.getElementById("audioPlayer");
	controls = document.getElementById("audioControls");
	metadata = document.getElementById("audioInfo");
	searchbar = document.getElementById("search");
	if (searchbar === null) {
		searchbar = document.getElementById("search hidden");
	}
	searchbox = document.getElementById("searchbox");
	searching = document.getElementById("searching");
	back = document.getElementById("back");
	readMore = document.getElementById("readMore");
	showLess = document.getElementById("showLess");
	moreText = document.getElementById("more");
	buttonVisibility = document.getElementById("visibility");
	notif = document.getElementById("notification");
}

render = (match, content) => {
	//Make sure we know our stuff first
	refreshElements()

	//Clear the page if we're rendering something
	if (match != null && content != null)
		clearPage();

	//Set and capture the page
	pageContent = content;
	if (match !== null) {
		pageCapture(match);
	}

	//Lastly, render it
	if (content !== '') {
		document.querySelector("#results").innerHTML = '<table>' + content + '</table>';
	} else {
		document.querySelector("#results").innerHTML = '';
	}
	refreshElements();
	navigo.updatePageLinks();

	//Set the new navigation buttons
	setNavButtons();
};

//Button navigation for embedded clients
function setNavButtons() {
	var btnBack = '<a href="/back" data-navigo><i class="fa-solid fa-arrow-left"></i></a>';
	var btnVisibility = '<a id="visibility" onclick="toggleVisibility()"><i class="fa-solid fa-eye';
	if (!visibility)
		btnVisibility += '-slash';
	btnVisibility += '"></i></a>';
	var btns = '';
	if (pageNum > 0 && visibility)
		btns += btnBack + ' ';
	btns += btnVisibility;
	/*
	if (pageHistory.length > 0 && pageNum < (pageHistory.length-1))
		btns += btnForward;
	*/

	//Lastly, render it
	document.querySelector("#nav").innerHTML = btns;
	refreshElements();
	navigo.updatePageLinks();
}

//Toggles the visibility of the search box and the audio player
function toggleVisibility() {
	if (visibility) {
		//console.log("Hiding elements");
		visibility = false;
		interruptNotification = true;
		buttonVisibility.innerHTML = '<i class="fa-solid fa-eye-slash"></i>';
		infobar.setAttribute("id", "infobar hidden");
		searchbar.setAttribute("id", "search hidden");

		//Destroy the child elements
		metadata.innerHTML = "";
		controls.innerHTML = "";
		timer.innerHTML = "";
		createdAudioPlayer = false;
		searchbar.innerHTML = "";
		createdSearchBar = false;
	} else {
		//console.log("Showing elements");
		visibility = true;
		interruptNotification = false;
		buttonVisibility.innerHTML = '<i class="fa-solid fa-eye"></i>';
		infobar.setAttribute("id", "infobar");
		searchbar.setAttribute("id", "search");

		//Create the child elements again
		createAudioPlayer();
		createSearchBar();
	}

	setNavButtons();
}

function clearPage() {
	//console.log("clearPage");
	refreshElements();
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

function pageCapture(match) {
	if (captureLock) {
		//console.log("Lock prevented page capture!");
		captureLock = false;
		return;
	}
	pageNum++;
	pageHistory.splice(pageNum, (pageHistory.length - pageNum - 1));
	if (match !== null) {
		lastPageUrl = match.url;
		if (match.queryString !== "")
			lastPageUrl += "?" + match.queryString;
	}
	var page = [lastPageUrl, pageObject, bgImg];
	pageHistory.push(page);
	//console.log("Captured page " + (pageNum+1) + "/" + pageHistory.length + "! " + lastPageUrl);
}

function pageRelease() {
	if (pageNum-1 < 0) {
		//console.log("No way backward!");
		return;
	}
	var oldPage = pageHistory[pageNum];
	if (oldPage == null || (oldPage[0].substring(0, 7) != "/stream" && oldPage[0].substring(0, 9) != "/download" && oldPage[0].substring(0, 6) != "/addqueue" && oldPage[0].substring(0, 11) != "/transcript")) {
		pageHistory.splice(pageNum, (pageHistory.length - pageNum));
		captureLock = true;
		pageNum--;
		var page = pageHistory[pageNum];
		//console.log("Releasing page " + (pageNum+1) + "/" + pageHistory.length + "! " + page[0]);
		pageImpose(page);
	}
}

function pageImpose(page) {
	if (page == null) {
		return;
	}
	navigo.navigate(page[0], {callHandler: true});
	lastPageUrl = page[0];
	pageObject = page[1];
	//render(null, page[1]);
	setBgImg(page[2]);
	navigo.updatePageLinks();
}

function pagePotato(match) {
	pageCapture(null);
	pageRelease();
}
