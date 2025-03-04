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
  /** @type {HTMLButtonElement} */
  const timeout = document.getElementById("timeout");

  var start = Date.now();
  id = setInterval(function () {
    /** @type {Number} */
    let delta = Date.now() - start;
    /** @type {Number} */
    let countDown = 5;
    countDown -= Math.floor(delta / 1000);

    timer.innerHTML = countDown;

    if (countDown <= 0) {
      timeout.click();
      clearInterval(id);
    }
  }, 1000);
});
