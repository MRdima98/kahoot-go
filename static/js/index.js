window.addEventListener("load", function () {
  const img = document.getElementById("test");

  if (img.naturalWidth < img.naturalHeight) {
    img.style.width = "600px";
  } else {
    img.style.width = "1800px";
  }
  //document.body.addEventListener("htmx:wsAfterMessage", function (event) {
  //  console.log("Message received:", event.detail.message);
  //  document
  //    .querySelector("#chat_messages")
  //    .insertAdjacentHTML("beforeend", event.detail.message);
  //});
});
