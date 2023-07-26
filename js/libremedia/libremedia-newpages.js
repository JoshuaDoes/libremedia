var cache = new Map();

async function objectGet(objectUri) {
	try {
		var obj = cache.get(objectUri);
		if (obj == null) {
			obj = await v1GetObject(objectUri);
			cache.set(objectUri, obj);
		}
		return objectCondense(obj);
	} catch (error) {
		console.error(`Failed to get ${objectUri}: ${error}`);
		throw error;
	}
}

async function objectRender(match) {
	if (match == null || match.params == null || match.params.uri == null)
		return;
	var objectUri = match.params.uri;

	try {
		var obj = await objectGet(objectUri);
		if (obj == null)
			return;

		var html = "";
		switch (obj.type) {
		case "search":
			html = renderSearch(obj);
			break;
		case "creator":
			html = renderCreator(obj);
			break;
		case "album":
			html = renderAlbum(obj);
			break;
		default:
			return;
		}

		if (html != "")
			render(match, html);
	} catch (error) {
		console.error(`Failed to render ${objectUri}: ${error}`);
		throw error;
	}
}

function renderSearch(obj) {
	if (obj == null)
		return;

	var html = "";
	if (obj.streams != null)
		pageObject = obj.streams; //Cache the streams list for the queue
		html += tblStreams(obj.streams);
	if (obj.creators != null)
		html += tblCreators(obj.creators);
	if (obj.albums != null)
		html += tblAlbums(obj.albums);
	return html;
}

function renderCreator(obj) {
	if (obj == null)
		return;

	var html = "";
	return html;
}

function renderAlbum(obj) {
	if (obj == null)
		return;

	var html = "";
	return html;
}

function renderLinkStream(obj) {
	if (obj == null)
		return '<div id="stream">' + textRefresh + '</div>';

	var html = '<div id="stream"><a href="/stream?uri=' + obj.uri + '" data-navigo>';
	if (obj.name != "") {
		html += obj.name;
	} else {
		html += textRefresh;
	}
	if (obj.explicit)
		html += ' ' + iconExplicit;
	return html + '</a></div>';
}

function renderLinkCreator(obj) {
	if (obj == null)
		return '<div id="creator">' + textRefresh + '</div>';

	var html = '<div id="creator"><a href="/creator?uri=' + obj.uri + '" data-navigo>';
	if (obj.name != "") {
		html += obj.name;
	} else {
		html += textRefresh;
	}
	return html + '</a></div>';
}

function renderLinkAlbum(obj) {
	if (obj == null)
		return '<div id="album">' + textRefresh + '</div>';

	var html = '<div id="album"><a href="/album?uri=' + obj.uri + '" data-navigo>';
	if (obj.name != "") {
		html += obj.name;
	} else {
		html += textRefresh;
	}
	if (obj.explicit)
		html += ' ' + iconExplicit;
	return html + '</a></div>';
}

function renderTableRowStream(obj) {
	if (obj == null)
		return;

	var html = renderLinkStream(obj);
	if (obj.creators != null) {
		html += '<br />' + renderLinkCreator(obj.creators[0]);
		for (let i = 1; i < obj.creators.length; i++) { //Loop doesn't execute unless more than one creator
			html += ', ' + renderLinkCreator(obj.creators[i]);
		}
	}
	if (obj.album != null)
		html += '<br />' + renderLinkAlbum(obj.album);
	html = `<td>${html}</td>`;
	html += '<td>' + secondsTimestamp(obj.duration) + '</td>';
	html += '<td>' + renderControlsStream(obj) + '</td>';
	return `<tr>${html}</tr>`;
}

function renderControlsStream(obj) {
	if (obj == null)
		return;

	var html = '<a href="/addqueue?uri=' + obj.uri + '" data-navigo>' + iconAddQueue + '</a>';
	html += ' <a href="/download?uri=' + obj.uri + '" data-navigo>' + iconDownload + '</a>'
	if (obj.transcript != null)
		html += ' <a href="/transcript?uri=' + stream.uri + '" data-navigo>' + iconTranscript + '</a>';
	return `<div id="controls">${html}</div>`;
}

function objectCondense(obj) {
	if (obj == null)
		return null;
	if (obj.object != null)
		obj = obj.object;
	return obj;
}