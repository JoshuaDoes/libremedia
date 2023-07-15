//Elements
var content; //Wrapper for everything
var infobar;
var timer;
var player;
var metadata;
var download;
var searchbar;
var searchbox;
var searching;
var back;
var readMore;
var showLess;
var moreText;
var buttonPrev;
var buttonPP;
var buttonNext;
var buttonDownload;
var buttonTranscript;
var buttonRepeat;
var buttonVisibility;
var notif;

var visibility = true; //Used when toggling visibility of the search box and audio player

//Playback management
var queue = []; //Holds a list of queued streams, starting with the user's queue, followed by the queue of the current page
var queueStart = -1; //The index of the first user-added stream in the queue
var queueEnd = -1; //The index of the last user-added stream in the queue
var queueLeft = 0; //The total amount of streams left to play before the end of the queue
var nowPlaying; //The stream currently loaded into the audio player
var nowPlayingTiming = []; //The current stream's transcript timings for seeking and following along
var shuffle = false;
var repeat = 0; //0=no repeat, 1=repeat queue, 2=repeat now playing
var lyricScrollerId; //Holds an id returned by setInterval, used to clear timer on page clear
var lastScrollY = -1; //The last recorded Y-axis scroll position, used to cancel auto-scroller
var lastLyric = -1; //The last recorded lyric that was auto-scrolled to, -1 means hasn't been scrolled and 0 means beginning of stream

//Search results
var query = "";
var previousQuery = "";
var delayTimer;

//Single page routing with navigo
var navigo;
var pageHistory = [];
var pageNum = -1;
var pageContent = "";
var pageObject = [];
var bgImg = "";

var captureLock = false; //Used to prevent captures on scripted pages that don't render any content
var lastPageUrl = "/"; //Used to be globally aware of the current page

var createdSearchBar = false;
var createdAudioPlayer = false;
var interruptNotification = false;

//Static icons and texts
var textRefresh = '<small>refresh to try again</small>';
var iconExplicit = '<i class="bi bi-explicit"></i>';
var iconSearching = '🔎';
var iconLoading = '<i class="fa-solid fa-music"></i>';
var iconDownload = '<i class="bi bi-file-earmark-arrow-down"></i>';
var iconAddQueue = '<i class="bi bi-music-note-list"></i>';
var iconTranscript = '<i class="bi bi-body-text"></i>';
var iconPrevious = '<i class="fa-solid fa-backward-step"></i>';
var iconNext = '<i class="fa-solid fa-forward-step"></i>';
var iconRepeat = '<i class="fa-solid fa-right-long"></i>';

$(document).ready(function() {
	console.log("Setting up libremedia...");
	refreshElements();
	createAudioPlayer();
	createSearchBar();
	refreshQuery();
	navigoResolve();
	console.log("Finished constructing libremedia instance!");
});