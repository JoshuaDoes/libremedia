/* API handlers */

function get(link) {
	return new Promise((resolve, reject) => {
		const HttpReq = new XMLHttpRequest();
		HttpReq.open("GET", link);
		HttpReq.onload = function() {
			if (HttpReq.status === 200) {
				resolve(HttpReq.responseText);
			} else {
				reject(new Error(HttpReq.statusText));
			}
		};
		HttpReq.onerror = function() {
			reject(new Error("Network error"));
		};
		HttpReq.send();
	});
}

async function getData(link) {
	try {
		const response = await get(link);
		return response;
	} catch (error) {
		console.error("Error fetching data:", error);
		throw error;
	}
}

async function getJson(link) {
	try {
		const jsonData = await getData(link);
		const jsonObj = JSON.parse(jsonData);
		return jsonObj;
	} catch (error) {
		console.error("Error parsing JSON:", error);
		throw error;
	}
}

async function v1GetBestMatch(query) {
	try {
		const result = await getJson("/v1/bestmatch:" + query);
		return result;
	} catch (error) {
		console.error("Error in v1GetBestMatch:", error);
		throw error;
	}
}

async function v1GetSearch(query) {
	try {
		const result = await getJson("/v1/search:" + query);
		return result;
	} catch (error) {
		console.error("Error in v1GetSearch:", error);
		throw error;
	}
}

async function v1GetObject(object) {
	try {
		const result = await getJson("/v1/" + object);
		return result;
	} catch (error) {
		console.error("Error in v1GetObject:", error);
		throw error;
	}
}

async function v1GetStream(object) {
	try {
		const result = await getJson("/v1/stream/" + object);
		return result;
	} catch (error) {
		console.error("Error in v1GetStream:", error);
		throw error;
	}
}

async function v1GetStreamBestMatch(query) {
	try {
		const result = await getJson("/v1/stream/bestmatch:" + query);
		return result;
	} catch (error) {
		console.error("Error in v1GetStreamBestMatch:", error);
		throw error;
	}
}