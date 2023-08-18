var menu = document.getElementsByClassName("main-menu")[0];
var anchors = menu.getElementsByTagName("a");
var trigger;

for (var i = 0; i < anchors.length; i++) {
  a = anchors[i];
  if (a.href.indexOf("#theme") >= 0) {
    trigger = a;
  }
}

theme = document.cookie.match("theme=([^;]+)");

if (theme) {
  mode = theme[1];

  if (mode == "night") {
    document.body.classList.add("night");
    document.body.classList.remove("day");
    trigger.innerText = "Day Mode";
  }

  if (mode == "day") {
    document.body.classList.add("day");
    document.body.classList.remove("night");
    trigger.innerText = "Night Mode";
  }
}

function themeSwitch(e) {
  if (e.target.href.indexOf("#theme") >= 0) {
    document.body.classList.toggle("night");

    if (document.body.classList.contains("night")) {
      e.target.innerText = "Day Mode";
      document.cookie = "theme=night;domain=.integralist.co.uk;path=/";
    } else {
      e.target.innerText = "Night Mode";
      document.cookie = "theme=day;domain=.integralist.co.uk;path=/";
    }
  }
}

var menu = document.getElementById("navmenu");

menu.addEventListener("click", themeSwitch);

///////////////////////////////////////////////////////////////////////////////
//
// The following script ensures all external links open in a new window.
//
///////////////////////////////////////////////////////////////////////////////

document.addEventListener('DOMContentLoaded', function() {
  // Get all anchor elements on the page (excepts those that are listing blog posts, like on the home page)
  // NOTE: The :not() doesn't look to work :-(
  var anchors = document.querySelectorAll('a[href]:not(.list-item)');

  // Loop through each anchor element
  anchors.forEach(function(anchor) {
    // Get the value of the "href" attribute
    var href = anchor.getAttribute('href');

    // Check if the link is external
    if (isExternalLink(href)) {
      // Add target="_blank" to open the link in a new tab
      anchor.setAttribute('target', '_blank');
      anchor.setAttribute('class', 'external-link');
    }
  });
});

// Function to check if a link is external
function isExternalLink(url) {
  // Define the allowed domains (integralist.co.uk and localhost)
  var allowedDomains = ['www.integralist.co.uk', 'localhost'];

  // Check if the URL is valid
  try {
    var hostname = new URL(url).hostname;

    // Check if the hostname is not in the allowed domains
    return allowedDomains.indexOf(hostname) === -1;
  } catch (error) {
    // Handle invalid URLs here (e.g., relative paths)
    return false;
  }
}
