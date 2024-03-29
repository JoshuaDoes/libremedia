//Elements
var btnPP;
var audio;
var timer;

//Icons
var play = 'PLAY';
var pause = 'PAUSE';
var loading = "LOADING";

//State
var ready = false;
var playing = false;
var init = false;

function audioInit() {
  btnPP = document.getElementById("btnPP");
  if (btnPP != null)
    btnPP.addEventListener("click", audioPP);
  audio = document.getElementById("audioPlayer");
  timer = document.getElementById("audioTimer");

  if (init) {
    return;
  }

  if (audio != null) {
    audio.addEventListener("loadstart", audioLoad);
    audio.addEventListener("canplay", audioReady);
    audio.addEventListener("pause", audioPause);
    audio.addEventListener("play", audioPlay);
    audio.addEventListener("playing", audioResume);
    audio.addEventListener("timeupdate", audioTime);
    audio.addEventListener("waiting", audioBuffer);
    audio.addEventListener("ended", audioEnd);
  }

  init = true;
};

function audioPP() {
  if (audio.paused) {
    audioPlay(null);
  } else {
    audioPause(null);
  }
}

function audioLoad(event) {
  ready = false;
  playing = false;
  if (btnPP == null) {
    return;
  }
  btnPP.innerHTML = loading;
}

function audioReady(event) {
  ready = true;
}

function audioPlay(event) {
  if (audio.src == "")
    return;
  if (!playing) {
    audio.play();
  }
  playing = true;
}

function audioPause(event) {
  audio.pause();
  if (btnPP == null) {
    return;
  }
  btnPP.innerHTML = play;
  playing = false;
}

function audioResume(event) {
  if (btnPP == null) {
    return;
  }
  btnPP.innerHTML = pause;
}

function audioTime(event) {
  if (timer.innerHTML == "" || timer.innerHTML == "Waiting to stream...") {
    return;
  }
  const pos = audio.currentTime;
  const len = audio.duration;
  if (len == NaN || len <= 0) {
    timer.innerHTML = "0:00 / 0:00";
    return;
  }
  timer.innerHTML = Math.floor(pos / 60) + ":" + Math.floor(pos % 60).toString().padStart(2, '0') + " / " + Math.floor(len / 60) + ":" + Math.floor(len % 60).toString().padStart(2, '0');
}

function audioBuffer(event) {
  if (btnPP == null) {
    return;
  }
  btnPP.innerHTML = loading;
}

function audioEnd(event) {
  playing = false;
  if (btnPP == null) {
    return;
  }
  btnPP.innerHTML = play;
}