function copy() {
  const el = document.getElementById("data");
  copyText.select();
  copyText.setSelectionRange(0, 99999); // For mobile devices
  navigator.clipboard.writeText(el.value);
}

window.onload = () => {
  console.log("ok");

  function clearAddErrors() {
    const els = [document.querySelector("textarea#text")];
    els.forEach((e) => {
      e.addEventListener("click", () => {
        const sEl = document.querySelector("p#error");
        if (sEl != null) {
          sEl.remove();
        }
      });
    });
  }

  document.body.addEventListener("htmx:load", function () {
    clearAddErrors();
  });
};
