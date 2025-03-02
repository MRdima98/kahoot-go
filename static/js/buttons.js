function disableButtons() {
  document.addEventListener("htmx:oobAfterSwap", function (event) {
    /**
     * @type {HTMLButtonElement[]}
     */
    const buttons = Array.from(document.getElementsByTagName("button"));
    buttons.forEach((el) => {
      console.log("loop for disables");
      el.disabled = true;
    });
  });
}
