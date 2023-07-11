function createSearchBar() {
	if (createdSearchBar) {
		return;
	}
	createdSearchBar = true;

	searchbar.innerHTML = "";
	searchbox = document.createElement("input");
	searchbox.setAttribute("id", "searchbox");
	searchbox.setAttribute("type", "text");
	searchbox.setAttribute("placeholder", "🔎 creator, stream, album ...");
	notif = document.createElement("div");
	notif.setAttribute("id", "notification");
	searching = document.createElement("div");
	searching.setAttribute("id", "searching");
	searchbar.appendChild(searchbox);
	searchbar.appendChild(notif);
	searchbar.appendChild(searching);

	$("#searchbox").keyup(function() {
		refreshQuery();
		if (previousQuery == query) {
			return;
		}
		previousQuery = query;
		if (query == "") {
			navigo.navigate("/");
		} else {
			navigo.navigate("search?q=" + query);
		}
	});
}

function refreshQuery() {
	query = $("#searchbox").val();
}