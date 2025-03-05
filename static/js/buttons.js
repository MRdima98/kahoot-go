function disableButtons() {
  document.addEventListener("htmx:oobAfterSwap", function () {
    /**
     * @type {HTMLButtonElement[]}
     */
    const buttons = Array.from(document.getElementsByTagName("button"));
    buttons.forEach((el) => {
      el.disabled = true;
    });
  });
}
