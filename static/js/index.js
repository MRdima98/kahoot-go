window.addEventListener("load", function () {
  /** @type {HTMLImageElement} */
  const img = document.getElementById("test");

  if (img.naturalWidth < img.naturalHeight) {
    img.style.width = "600px";
  } else {
    img.style.width = "1800px";
  }

  /** @type {HTMLDivElement} */
  const timer = document.getElementById("timer");

  var start = Date.now();
  setInterval(function () {
    /** @type {Number} */
    let delta = Date.now() - start;
    /** @type {Number} */
    let countDown = 5;
    countDown -= Math.floor(delta / 1000);

    if (countDown <= 0) {
      return;
    }
    timer.innerHTML = countDown;
  }, 1000);
});
