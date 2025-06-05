/** @type {Number} */
let timer_id;
let countDown = 10;

document.addEventListener("htmx:wsAfterMessage", function(event) {
  countDown = 10;
  clearInterval(timer_id)
  triggerTimer();
  document.gameSocket = event.detail.socketWrapper;
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
  /** @type {HTMLButtonElement} */
  if (timer_id != null) {
    clearInterval(timer_id);
  }
  let timer = null;

  timer_id = setInterval(function() {
    countDown -= 1;
    if (!timer) {
      timer = document.getElementById("timer");
    }
    timer.innerHTML = countDown;

    if (countDown <= 0) {
      clearInterval(timer_id);
      console.log(document.gameSocket)
      document.gameSocket.send('timeout')
    }
  }, 1000);
}
