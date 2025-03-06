/** @type {Function} */
let disableButtons;
window.addEventListener("load", function () {
  disableButtons = () => {
    const buttons = Array.from(document.getElementsByTagName("button"));
    buttons.forEach((el) => {
      el.disabled = true;
    });
  };
});
