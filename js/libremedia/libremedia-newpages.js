var cache = new Map();

async function objectGet(objectUri) {
	try {
		var obj = cache.get(objectUri);
		if (obj == null) {
			obj = await v1GetObject(objectUri);
			cache.set(objectUri, obj);
		}
		return obj;
	} catch (error) {
		console.error(`Failed to get ${objectUri}: ${error}`);
		throw error;
	}
}

async function objectRender(match) {
	console.log(match);
	if (match == null || match.params == null || match.params.uri == null)
		return;
	var objectUri = match.params.uri;
	console.log(objectUri);

	try {
		var obj = await objectGet(objectUri);
		if (obj == null)
			return;
		console.log(obj);

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
			console.error('Unknown object type ' + obj.type);
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
	obj = objectCondense(obj);

	var html = "";
	if (obj.streams != null)
		pageObject = obj.streams; //Cache the streams list for the queue
		html += renderTableListStreams("Streams", obj.streams);
	if (obj.creators != null)
		html += renderTableListCreators("Creators", obj.creators);
	if (obj.albums != null)
		html += renderTableListAlbums("Albums", obj.albums);
	return html;
}

function renderLinkStream(obj) {
	if (obj == null)
		return '<div id="stream">' + textRefresh + '</div>';
	obj = objectCondense(obj);

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

function renderTableRowStream(stream) {
	if (stream == null)
		return;
	stream = objectCondense(stream);

	var html = renderLinkStream(stream);
	if (stream.creators != null) {
		html += '<br />' + renderLinkCreator(stream.creators[0]);
		for (let i = 1; i < stream.creators.length; i++) { //Loop doesn't execute unless more than one creator
			html += renderLinkCreator(stream.creators[i]);
		}
	}
	if (stream.album != null)
		html += '<br />' + renderLinkAlbum(stream.album);
	html = `<td>${html}</td>`;
	html += '<td>' + secondsTimestamp(stream.duration) + '</td>';
	html += '<td>' + renderControlsStream(stream) + '</td>';
	return `<tr>${html}</tr>`;
}

function renderTableListStreams(label, streams) {
	var html = `<tr><th>${label}</th><th>🕑</th></tr>`;
	for (let i = 0; i < streams.length; i++) {
		html += renderTableRowStream(streams[i]);
	}
	return html;
}

function renderControlsStream(obj) {
	if (obj == null)
		return;
	obj = objectCondense(obj);

	var html = '<a href="/addqueue?uri=' + obj.uri + '" data-navigo>' + iconAddQueue + '</a>';
	html += ' <a href="/download?uri=' + obj.uri + '" data-navigo>' + iconDownload + '</a>'
	if (obj.transcript != null)
		html += ' <a href="/transcript?uri=' + obj.uri + '" data-navigo>' + iconTranscript + '</a>';
	return `<div id="controls">${html}</div>`;
}

function renderCreator(obj) {
	if (obj == null)
		return;
	obj = objectCondense(obj);

	var html = renderTableRowCreator(obj);
	if (obj.description != null && obj.description.length > 0)
		html += renderTableRowCreatorBio(obj.description);
	if (obj.genres != null && obj.genres.length > 0)
		html += renderTableRowCreatorGenres(obj.genres);
	if (obj.topStreams != null && obj.topStreams.length > 0)
		html += renderTableListStreams("Top Streams", obj.topStreams);
	if (obj.albums != null && obj.albums.length > 0)
		html += renderTableListAlbums("Albums", obj.albums);
	if (obj.singles != null && obj.singles.length > 0)
		html += renderTableListAlbums("Singles & EPs", obj.singles);
	if (obj.appearances != null && obj.appearances.length > 0)
		html += renderTableListAlbums("Album Features", obj.appearances);
	if (obj.related != null && obj.related.length > 0)
		html += renderTableListCreators("Related Creators", obj.related);
	return html;
}

function renderLinkCreator(obj) {
	if (obj == null)
		return '<div id="creator">' + textRefresh + '</div>';
	obj = objectCondense(obj);

	var html = '<div id="creator"><a href="/creator?uri=' + obj.uri + '" data-navigo>';
	if (obj.name != "") {
		html += obj.name;
	} else {
		html += textRefresh;
	}
	return html + '</a></div>';
}

function renderTableRowCreator(creator) {
	if (creator == null)
		return;
	creator = objectCondense(creator);

	var html = renderLinkCreator(creator);
	html = `<td>${html}</td>`;
	return `<tr>${html}</tr>`;
}

function renderTableListCreators(label, creators) {
	var html = `<tr><th>${label}</th></tr>`;
	for (let i = 0; i < creators.length; i++) {
		html += renderTableRowCreator(creators[i]);
	}
	return html;
}

function renderTableRowCreatorBio(bio) {
	bio = bio
		.replace(/\r\n|\r|\n/gim, '<br />') //Linebreaks
		.replace(/\[([^\[]+)\](\(([^)]*))\)/gim, '<a href="$3" data-navigo>$1</a>'); //Anchor tags
	var split = bio.split(' ');
	var bioLen = 15;

	var html = '';
	if (split.length > bioLen) {
		var bioView = split.slice(0, bioLen).join(' ');
		html += '<p id="readMore" onclick="textExpander()">' + bioView + ' ...<br /><small>(tap to show more)</small></p>';
		html += '<span id="more"><p id="showLess" onclick="textExpander()">' + bio + '<br /><small>(tap again to show less)</small></p></span>';
	} else {
		html += `<p>${bio}</p>`;
	}
	return `<tr><td>${html}</td></tr>`;
}

function renderTableRowCreatorGenres(genres) {
	if (genres == null || genres.length == 0)
		return;

	var html = genres[0];
	for (let i = 1; i < genres.length; i++) {
		html += ', ' + genres[i];
	}
	return `<tr><td>${html}</td></tr>`;
}

function renderAlbum(obj) {
	if (obj == null)
		return;
	obj = objectCondense(obj);

	var html = renderTableRowAlbum(obj);
	for (let disc = 0; disc < obj.discs.length; disc++) {
		html += renderTableListStreams("Disc " + (disc+1), obj.discs[disc].streams);
	}
	return html;
}

function renderLinkAlbum(obj) {
	if (obj == null)
		return '<div id="album">' + textRefresh + '</div>';
	obj = objectCondense(obj);

	var html = '<div id="album"><a href="/album?uri=' + obj.uri + '" data-navigo>';
	if (obj.name != "") {
		html += obj.name;
	} else {
		html += textRefresh;
	}
	if (obj.explicit)
		html += ' ' + iconExplicit;
	if (obj.datetime != null)
		html += '<br />(' + obj.datetime + ')';
	return html + '</a></div>';
}

function renderTableRowAlbum(album) {
	if (album == null)
		return;
	album = objectCondense(album);

	var html = renderLinkAlbum(album);
	if (album.creators != null) {
		html == '<br />' + renderLinkCreator(album.creators[0]);
		for (let i = 1; i < album.creators.length; i++) { //Loop doesn't execute unless more than one creator
			html += renderLinkCreator(album.creators[i]);
		}
	}
	html = `<td>${html}</td>`;
	return `<tr>${html}</tr>`;
}

function renderTableListAlbums(label, albums) {
	var html = `<tr><th>${label}</th></tr>`;
	for (let i = 0; i < albums.length; i++) {
		html += renderTableRowAlbum(albums[i]);
	}
	return html;
}

function objectCondense(obj) {
	if (obj == null)
		return null;
	if (obj.object != null)
		obj = obj.object;
	return obj;
}