function disableButtons() {
  document.addEventListener("htmx:oobAfterSwap", function () {
    /**
     * @type {HTMLButtonElement[]}
     */
    const buttons = Array.from(document.getElementsByTagName("button"));
    buttons.forEach((el) => {
      console.log("loop for disables");
      el.disabled = true;
    });
  });

  document.addEventListener("htmx:wsOpen", function () {
    /** @type {HTMLButtonElement} */
    const player = document.getElementById("player");
    console.log("player");
    player.click();
  });
}
