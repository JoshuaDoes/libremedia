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
	element.style.display = "inline";
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