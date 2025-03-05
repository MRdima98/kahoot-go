window.addEventListener("load", function () {
  triggerTimer();
  observeMutation();
});

//window.addEventListener("htmx:afterProcessNode", function () {
//  triggerTimer();
//});

function fixDim() {
  const img = document.getElementById("picture");

  if (img.naturalWidth < img.naturalHeight) {
    img.style.width = "600px";
  } else {
    img.style.width = "1800px";
  }
}

function triggerTimer() {
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
}

function observeMutation() {
  /** @type {HTMLDivElement} */
  let targetNode = document.body;
  const config = { childList: true, subtree: true };

  const callback = () => {
    /** @type {HTMLDivElement} */
    let timer = document.getElementById("timer");

    if (timer != null && Number(timer.innerHTML) === 30) {
      console.log("gotem");
      triggerTimer();
    }
  };

  const observer = new MutationObserver(callback);
  observer.observe(targetNode, config);
}
