window.addEventListener("load", function () {
  const img = document.getElementById("test");

  if (img.naturalWidth < img.naturalHeight) {
    img.style.width = "150px";
  } else {
    img.style.width = "300px";
  }

  console.log("Width:", img.naturalWidth);
  console.log("Height:", img.naturalHeight);
});
