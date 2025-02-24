window.addEventListener("load", function () {
  const img = document.getElementById("test");

  if (img != null) {
    if (img.naturalWidth < img.naturalHeight) {
      img.style.width = "600px";
    } else {
      img.style.width = "1800px";
    }
  }
});
