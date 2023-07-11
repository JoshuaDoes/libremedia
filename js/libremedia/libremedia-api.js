/* API handlers */

function Get(link) {
	var HttpReq = new XMLHttpRequest();
	HttpReq.open("GET", link, false);
	HttpReq.send(null);
	return HttpReq;
}
function GetData(link) {
	return Get(link).response;
}
function GetText(link) {
	return Get(link).responseText;
}
function GetJson(link) {
	var jsonData = GetData(link);
	var jsonObj = JSON.parse(jsonData);
	return jsonObj;
}

function v1GetBestMatch(query) {
	return GetJson('/v1/bestmatch:' + query);
}
function v1GetSearch(query) {
	return GetJson('/v1/search:' + query);
}
function v1GetObject(object) {
	return GetJson('/v1/' + object);
}
function v1GetStream(object) {
	return GetJson('/v1/stream/' + object);
}
function v1GetStreamBestMatch(query) {
	return GetJson('/v1/stream/bestmatch:' + query);
}