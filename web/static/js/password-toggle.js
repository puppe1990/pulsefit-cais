document.addEventListener("click", (e) => {
  const btn = e.target.closest(".password-toggle");
  if (!btn) return;

  const input = btn.parentElement?.querySelector("input");
  if (!input) return;

  const show = input.type === "password";
  input.type = show ? "text" : "password";

  const showIcon = btn.querySelector(".password-toggle-show");
  const hideIcon = btn.querySelector(".password-toggle-hide");
  showIcon?.classList.toggle("hidden", show);
  hideIcon?.classList.toggle("hidden", !show);
  btn.setAttribute("aria-label", show ? "Hide password" : "Show password");
});