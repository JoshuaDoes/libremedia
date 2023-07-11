function downloadStream(match) {
	if (match.params == null) {
		pagePotato(match);
		return;
	}
	var uri = match.params.uri;
	//console.log("Download: " + uri);
	window.open("/v1/download/" + uri, "_blank");
	pagePotato(match);
}
function downloadAlbum(albumURI) {
	//console.log("Not implemented yet! TODO: Download " + albumURI + " as ZIP");
}
function downloadDiscography(creatorURI) {
	//console.log("Not implemented yet! TODO: Download " + creatorURI + " as ZIP");
}
