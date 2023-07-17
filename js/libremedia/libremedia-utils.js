//Shamelessly copied from https://www.w3schools.com/js/js_cookies.asp
function getCookie(cname) {
	let name = cname + "=";
	let ca = document.cookie.split(';');
	for(let i = 0; i < ca.length; i++) {
		let c = ca[i];
		while (c.charAt(0) == ' ') {
			c = c.substring(1);
		}
		if (c.indexOf(name) == 0) {
			const val = c.substring(name.length, c.length);
			//console.log("Got cookie: " + cname + "=" + val);
			return val;
		}
	}
	//console.log("No cookie for " + cname);
	return "";
}

function setCookie(cname, cvalue, exdays) {
	const d = new Date();
	d.setTime(d.getTime() + (exdays * 24 * 60 * 60 * 1000));
	let expires = "expires="+d.toUTCString();
	document.cookie = cname + "=" + cvalue + ";" + expires + ";path=/;SameSite=strict";
	//console.log("Updated cookies: " + document.cookie);
}
//---------------------------------------------------------------------

//Google Chrome has a limit of 400 days, so that's what we'll use
function setPermaCookie(cname, cvalue) {
	setCookie(cname, cvalue, 400);
	//console.log("Set permanent cookie: " + cname + "=" + cvalue);
}

function textExpander() {
	if (elementHidden(readMore)) {
		elementShow(readMore);
		elementHide(showLess);
		elementHide(moreText);
	} else {
		elementHide(readMore);
		elementShow(showLess);
		elementShow(moreText);
	}
}

function elementHide(element) {
	element.style.display = "none";
}

function elementShow(element) {
	element.style.display = "block";
}

function elementHidden(element) {
	return (element.style.display === "none");
}

function elementVisible(element) {
	return !elementHidden(element);
}

//Reset the scroll position to the top left of the document
function resetScroll() {
	if (window.scrollY) {
		window.scroll(0, 0);
	}
}

function iconProvider(provider) {
	switch (provider) {
		case "spotify":
			return iconProviderSpotify;
		case "tidal":
			return iconProviderTidal;
	}
	return iconProviderLocal;
}

function sanitizeWhitespace(input) {
	var output = $('<span>').text(input).html();
	output = output.replace(" ", "+");
	return output;
}

function secondsTimestamp(seconds) {
	// Hours, minutes and seconds
	var hrs = ~~(seconds / 3600);
	var mins = ~~((seconds % 3600) / 60);
	var secs = ~~seconds % 60;

	// Output like "1:01" or "4:03:59" or "123:03:59"
	var ret = "";
	if (hrs > 0) {
		ret += "" + hrs + ":" + (mins < 10 ? "0" : "");
	}
	ret += "" + mins + ":" + (secs < 10 ? "0" : "");
	ret += "" + secs;
	return ret;
}