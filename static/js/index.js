/** @type {Number} */
let timer_id;
window.addEventListener("load", function () {
  triggerTimer();
  observeMutation();
});

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
  if (timer_id != null) {
    clearInterval(timer_id);
  }

  var start = Date.now();
  timer_id = setInterval(function () {
    /** @type {Number} */
    let delta = Date.now() - start;
    /** @type {Number} */
    let countDown = 60;
    countDown -= Math.floor(delta / 1000);

    timer.innerHTML = countDown;

    if (countDown <= 0) {
      timeout.click();
      clearInterval(timer_id);
    }
  }, 1000);
}

// function observeMutation() {
//   /** @type {HTMLDivElement} */
//   let targetNode = document.body;
//   const config = { childList: true, subtree: true };
//
//   const callback = () => {
//     /** @type {HTMLDivElement} */
//     let timer = document.getElementById("timer");
//
//     if (timer != null && Number(timer.innerHTML) === 30) {
//       triggerTimer();
//     }
//   };
//
//   const observer = new MutationObserver(callback);
//   observer.observe(targetNode, config);
// }
