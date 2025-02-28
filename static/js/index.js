window.addEventListener("load", function () {
  const img = document.getElementById("test");

  if (img != null) {
    if (img.naturalWidth < img.naturalHeight) {
      img.style.width = "600px";
    } else {
      img.style.width = "1800px";
    }
  }

  document.addEventListener("htmx:oobAfterSwap", function (event) {
    console.log(event);
    //if (event.target.id === "UI" || event.target.di === "Menu") {
    //  return;
    //}

    const buttons = document.getElementsByTagName("button");
    Array.from(buttons).forEach((el) => {
      el.disabled = true;
    });
  });
});
