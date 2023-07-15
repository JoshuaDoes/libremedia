function tblStream(provider, stream) {
	//console.log(stream);
	
	var html = '<td><div id="stream"><a href="/stream?uri=' + stream.uri + '" data-navigo>';
	if (stream.name !== "") {
		html += stream.name;
	} else {
		html += refresh;
	}
	if (stream.explicit) {
		html += ' ' + explicit;
	}
	html += '</a></div><br />';
	if (stream.creators != null) {
		html += '<div id="creator"><a href="/creator?uri=' + stream.creators[0].object.uri + '" data-navigo>' + stream.creators[0].object.name + '</a></div>';
	} else {
		html += refresh;
	}
	html += '<br />';
	if (stream.album != null) {
		html += '<div id="album"><a href="/album?uri=' + stream.album.object.uri + '" data-navigo>' + stream.album.object.name + '</a></div>';
	} else {
		html += refresh;
	}
	html += '</td><td>' + secondsTimestamp(stream.duration) + '<br /><div id="controls">';
	if (stream.transcript != null) {
		html += '<a href="/transcript?uri=' + stream.uri + '" data-navigo><i class="fa-solid fa-music"></i></a> ';
	}
	html += '<a href="/download?uri=' + stream.uri + '" data-navigo><i class="fa-solid fa-download"></i></a> ';
	html += '<a href="/addqueue?uri=' + stream.uri + '" data-navigo><i class="fa-solid fa-plus"></i></a></div></td>';

	if (stream.album.object.artworks != null) {
		var selbg = stream.album.object.artworks.length-1;
		if (selbg > 4) {
			selbg = 4;
		}
		const bestbg = stream.album.object.artworks[selbg];
		html = '<tr onmouseenter="setBgImg(\'' + bestbg.url + '\')" onmouseleave="resetBgImg()">' + html + '</tr>';
	} else {
		html = '<tr>' + html + '</tr>';
	}

	return html;
}
function tblStreams(streams) {
	html = '<tr><th>Streams</th><th>🕑</th></tr>'
	for (let i = 0; i < streams.length; i++) {
		html += tblStream(streams[i].provider, streams[i].object);
	}
	return html;
}

function tblCreator(provider, creator) {
	//console.log(creator);
	var html = "";
	html += '<td colspan="5"><div id="creator"><a href="/creator?uri=' + creator.uri + '" data-navigo>' + creator.name + '</a></div></td>';

	if (creator.artworks != null) {
		var selbg = creator.artworks.length-1;
		if (selbg > 4) {
			selbg = 4;
		}
		const bestbg = creator.artworks[selbg];
		html = '<tr onmouseenter="setBgImg(\'' + bestbg.url + '\')" onmouseleave="resetBgImg()">' + html + '</tr>';
	} else {
		html = '<tr>' + html + '</tr>';
	}

	return html;
}
function tblCreators(creators) {
	var html = '<tr><th colspan="5">Creators</th></tr>';
	for (let i = 0; i < creators.length; i++) {
		const creator = creators[i];
		html += tblCreator(creator.provider, creator.object);
	}
	return html;
}
function tblRelated(creators) {
	var html = '<tr><th colspan="5">Related</th></tr>';
	for (let i = 0; i < creators.length; i++) {
		const creator = creators[i];
		html += tblCreator(creator.provider, creator.object);
	}
	return html;
}
function tblStreamTop(provider, stream) {
	//console.log(stream);
	var refresh = '<small>refresh to try again</small>';
	var html = '<td><div id="stream"><a href="/stream?uri=' + stream.uri + '" data-navigo>';
	if (stream.name !== "") {
		html += stream.name;
	} else {
		html += refresh;
	}
	if (stream.explicit) {
		html += ' ' + explicit;
	}
	html += '</a></div><br />';
	if (stream.album != null) {
		html += '<div id="album"><a href="/album?uri=' + stream.album.object.uri + '" data-navigo>' + stream.album.object.name + '</a></div>';
	} else {
		html += refresh;
	}
	html += '</td><td>' + secondsTimestamp(stream.duration) + '<br /><div id="controls">';
	if (stream.transcript != null) {
		html += '<a href="/transcript?uri=' + stream.uri + '" data-navigo><i class="fa-solid fa-music"></i></a> ';
	}
	html += '<a href="/download?uri=' + stream.uri + '" data-navigo><i class="fa-solid fa-download"></i></a> ';
	html += '<a href="/addqueue?uri=' + stream.uri + '" data-navigo><i class="fa-solid fa-plus"></i></a></div></td>';

	if (stream.album.object.artworks != null) {
		var selbg = stream.album.object.artworks.length-1;
		if (selbg > 4) {
			selbg = 4;
		}
		const bestbg = stream.album.object.artworks[selbg];
		html = '<tr onmouseenter="setBgImg(\'' + bestbg.url + '\')" onmouseleave="resetBgImg()">' + html + '</tr>';
	} else {
		html = '<tr>' + html + '</tr>';
	}

	return html;
}
function tblStreamsTop(streams) {
	var html = '<tr><th>Top Streams</th><th>🕑</th></tr>';
	for (let i = 0; i < streams.length; i++) {
		html += tblStreamTop(streams[i].provider, streams[i].object);
	}
	return html;
}

function tblAlbum(provider, album) {
	//console.log(album);
	var html = '';
	html += '<td colspan="5"><div id="album"><a href="/album?uri=' + album.uri + '" data-navigo>' + album.name;
	if (album.explicit) {
		html += ' ' + explicit;
	}
	html += '</a></div></td>';

	if (album.artworks != null) {
		var selbg = album.artworks.length-1;
		if (selbg > 4) {
			selbg = 4;
		}
		const bestbg = album.artworks[selbg];
		html = '<tr onmouseenter="setBgImg(\'' + bestbg.url + '\')" onmouseleave="resetBgImg()">' + html + '</tr>';
	} else {
		html = '<tr>' + html + '</tr>';
	}

	return html;
}
function tblAlbums(albums) {
	var html = '<tr><th colspan="5">Albums</th></tr>';
	for (let i = 0; i < albums.length; i++) {
		const album = albums[i];
		html += tblAlbum(album.provider, album.object);
	}
	return html;
}
function tblSingles(albums) {
	var html = '<tr><th colspan="5">Singles & EPs</th></tr>';
	for (let i = 0; i < albums.length; i++) {
		const album = albums[i];
		html += tblAlbum(album.provider, album.object);
	}
	return html;
}
function tblAppearances(albums) {
	var html = '<tr><th colspan="5">Appears On</th></tr>';
	for (let i = 0; i < albums.length; i++) {
		const album = albums[i];
		html += tblAlbum(album.provider, album.object);
	}
	return html;
}
function tblStreamAlbum(provider, stream, number) {
	var refresh = '<small>refresh to try again</small>';
	var html = '<td><div id="stream"><a href="/stream?uri=' + stream.uri + '" data-navigo>';
	if (stream.name !== "") {
		html += stream.name;
	} else {
		html += refresh;
	}
	if (stream.explicit) {
		html += ' ' + explicit;
	}
	html += '</a></div></td><td>' + secondsTimestamp(stream.duration) + '<br /><div id="controls">';
	if (stream.transcript != null) {
		html += '<a href="/transcript?uri=' + stream.uri + '" data-navigo><i class="fa-solid fa-music"></i></a> ';
	}
	html += '<a href="/download?uri=' + stream.uri + '" data-navigo><i class="fa-solid fa-download"></i></a> ';
	html += '<a href="/addqueue?uri=' + stream.uri + '" data-navigo><i class="fa-solid fa-plus"></i></a></div></td>';

	if (stream.album.object.artworks != null) {
		var selbg = stream.album.object.artworks.length-1;
		if (selbg > 4) {
			selbg = 4;
		}
		const bestbg = stream.album.object.artworks[selbg];
		html = '<tr onmouseenter="setBgImg(\'' + bestbg.url + '\')" onmouseleave="resetBgImg()">' + html + '</tr>';
	} else {
		html = '<tr>' + html + '</tr>';
	}

	return html;
}