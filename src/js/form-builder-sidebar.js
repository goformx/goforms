// Form Builder Sidebar Toggle
document.addEventListener("DOMContentLoaded", () => {
  const sidebar = document.querySelector(".builder-sidebar");

  // Create and append toggle button if it doesn't exist
  if (!document.querySelector(".builder-sidebar-toggle")) {
    const toggleButton = document.createElement("button");
    toggleButton.className = "builder-sidebar-toggle";
    toggleButton.innerHTML = '<i class="bi bi-list"></i>';
    toggleButton.setAttribute("aria-label", "Toggle form components sidebar");
    document.body.appendChild(toggleButton);

    // Toggle sidebar on button click
    toggleButton.addEventListener("click", () => {
      sidebar.classList.toggle("is-open");
    });
  }

  // Close sidebar when clicking outside on mobile
  document.addEventListener("click", (e) => {
    const isMobile = window.innerWidth < 768;
    const isClickOutsideSidebar = !sidebar.contains(e.target);
    const isNotToggleButton = !e.target.closest(".builder-sidebar-toggle");

    if (
      isMobile &&
      isClickOutsideSidebar &&
      isNotToggleButton &&
      sidebar.classList.contains("is-open")
    ) {
      sidebar.classList.remove("is-open");
    }
  });

  // Handle window resize
  let resizeTimeout;
  window.addEventListener("resize", () => {
    clearTimeout(resizeTimeout);
    resizeTimeout = setTimeout(() => {
      if (window.innerWidth >= 768) {
        sidebar.classList.remove("is-open");
      }
    }, 250);
  });
});
