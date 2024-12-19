<!DOCTYPE HTML>
<!--
	Editorial by HTML5 UP
	html5up.net | @ajlkn
	Free for personal and commercial use under the CCA 3.0 license (html5up.net/license)
-->
<html>
  <head>
    <title>Integralist</title>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1, user-scalable=no" />
    <link rel="stylesheet" href="../../assets/css/main.css" />
  </head>
  <body class="is-preload">
		<button id="backToTop">↑ Back to Top</button>
    <!-- Wrapper -->
    <div id="wrapper">
      <!-- Main -->
      <div id="main">
        <div class="inner">
          <!-- Header -->
					<header id="header">
						<a href="../../index.html" class="logo"><strong>Home</strong></a>
							<ul class="icons">
							<!--<li><a href="https://x.com/integralist" class="icon brands fa-twitter" target="blank"><span class="label">Twitter</span></a></li>-->
							<!--<li><a href="https://instagram.com/wwfsuperstarsofwrestling" class="icon brands fa-instagram" target="blank"><span class="label">Instagram</span></a></li>-->
						</ul>
					</header>

          <!-- Content -->
          <section>
						<!--
            <header class="main">
              <h1>Royal Rumble 1989</h1>
              <p>A Bizarre Spectacle Where the Madness Multiplies</p>
            </header>
            <span class="image main"><img src="../images/rumble-89.jpg" alt="" /></span>
						-->
						{INSERT_MAIN}
          </section>
        </div>
      </div>
      <!-- Sidebar -->
<div id="sidebar">
  <div class="inner">
    <!-- Search -->
    <!--<section id="search" class="alt">-->
    <!--  <form method="post" action="#">-->
    <!--    <input type="text" name="query" id="query" placeholder="Search" />-->
    <!--  </form>-->
    <!--</section>-->
    <!-- Menu -->
    <nav id="menu">
      <header class="major">
        <h2>Menu</h2>
      </header>
      <ul>
        <li><a href="../../index.html">Home</a></li>
        <!--<li><a href="../resume/index.html">Resume</a></li>-->
				{INSERT_NAV}
      </ul>
    </nav>
    <!-- Section -->
		<!--
    <section>
      <header class="major">
        <h2>Highlights</h2>
      </header>
      <div class="mini-posts">
        <article>
          <a href="ppv-survivorseries-88.html" class="image"><img src="../images/survivor-88-index.jpg" alt="" /></a>
          <p>Get ready for the ultimate showdown as Survivor Series 1988 brings non-stop tag team action, fierce rivalries, and unforgettable battles between the biggest WWF superstars!</p>
        </article>
        <article>
          <a href="ppv-royalrumble-89.html" class="image"><img src="../images/rumble-89-index.jpg" alt="" /></a>
          <p>Royal Rumble 1989 unleashes chaos with 30 superstars battling for glory in an unforgettable over-the-top-rope showdown!</p>
        </article>
        <article>
          <a href="ppv-summerslam-88.html" class="image"><img src="../images/slam-88-index.jpg" alt="" /></a>
          <p>SummerSlam 1988 delivers explosive action with iconic matchups, as the WWF's biggest stars collide in the hottest event of the summer!</p>
        </article>
      </div>
    </section>
		-->
    <!-- Section -->
    <!--<section>-->
    <!--  <header class="major">-->
    <!--    <h2>Get in touch</h2>-->
    <!--  </header>-->
    <!--  <p>Sed varius enim lorem ullamcorper dolore aliquam aenean ornare velit lacus, ac varius enim lorem ullamcorper dolore. Proin sed aliquam facilisis ante interdum. Sed nulla amet lorem feugiat tempus aliquam.</p>-->
    <!--  <ul class="contact">-->
    <!--    <li class="icon solid fa-envelope"><a href="#">information@untitled.tld</a></li>-->
    <!--    <li class="icon solid fa-phone">(000) 000-0000</li>-->
    <!--    <li class="icon solid fa-home">1234 Somewhere Road #8254<br />-->
    <!--      Nashville, TN 00000-0000</li>-->
    <!--  </ul>-->
    <!--</section>-->
    <!-- Footer -->
    <footer id="footer">
      <p class="copyright">&copy; Integralist. All rights reserved.</p>
      <p class="copyright small">Demo Images: <a href="https://unsplash.com">Unsplash</a>. Design: <a href="https://html5up.net">HTML5 UP</a>.</p>
    </footer>
  </div>
</div>

    </div>
    <!-- Scripts -->
    <script src="../../assets/js/jquery.min.js"></script>
    <script src="../../assets/js/browser.min.js"></script>
    <script src="../../assets/js/breakpoints.min.js"></script>
    <script src="../../assets/js/util.js"></script>
    <script src="../../assets/js/main.js"></script>

		<!-- The following script is for handling the automatic TOC generated by `make build` -->
		<script>
		// Get references to the elements
		const nav = document.querySelector('#main nav');
		const h1 = document.querySelector('#main h1');

		// Move the `nav` element to be underneath the `h1`
		h1.insertAdjacentElement('afterend', nav);

		// Hide the `nav` element by default using inline styles
		nav.style.display = 'none';

		// Create a new `h2` element with the text "TOC"
		const toc = document.createElement('h2');
		toc.textContent = 'TOC';
		toc.className = "toc"

		// Add inline styles to make the `h2` look clickable
		// DISABLED: done with className
		//
		// toc.style.cursor = 'pointer';
		// toc.style.color = 'blue';
		// toc.style.textDecoration = 'underline';

		// Add a click event listener to toggle the visibility of the `nav`
		toc.addEventListener('click', () => {
				nav.style.display = nav.style.display === 'none' ? 'block' : 'none';
		});

		// Insert the `h2` element above the `nav`
		nav.insertAdjacentElement('beforebegin', toc);
		</script>

		<!-- The following script highlights the current page in the side nav -->
		<script>
		// Get the current page's URL path and normalize it
    const currentUrl = window.location.pathname;
    const normalizedCurrentUrl = currentUrl
        .replace(/.*\/(pages|posts)\//, '/$1/') // Ensure leading slash and extract from `pages/` or `posts/`
        .replace(/index\.html$/, ''); // Remove `index.html` suffix

    // Select all menu links
    const links = document.querySelectorAll('#menu ul li a');

    let matchedParentSpan = null;

    links.forEach(link => {
        // Normalize the link's href for comparison
        const normalizedHref = link.getAttribute('href')
            .replace(/^(\.\.\/)+/, '/') // Convert `../../` to `/` for consistency
            .replace(/index\.html$/, ''); // Remove `index.html` suffix

        // Check if the normalized href matches the normalized current URL
        if (normalizedHref === normalizedCurrentUrl) {
            // Add the inline style to the matching link
            link.style.color = 'black';

            // Find the parent span with the class 'opener'
            matchedParentSpan = link.closest('ul').previousElementSibling;
        }
    });

    // If a matching parent span was found, add the 'active' class
    if (matchedParentSpan && matchedParentSpan.classList.contains('opener')) {
        matchedParentSpan.classList.add('active');
    }
		</script>

		<!-- The following script handles "back to top" functionality -->
		<script>
    // Create a reference to the button
    const backToTopButton = document.getElementById('backToTop');

    // Show button when scrolled down a bit
    window.addEventListener('scroll', () => {
      if (window.scrollY > 300) {
        backToTopButton.style.display = 'block';
      } else {
        backToTopButton.style.display = 'none';
      }
    });

    // Add a click event listener to scroll to the top
    backToTopButton.addEventListener('click', () => {
      window.scrollTo({
        top: 0,
        behavior: 'smooth'
      });
    });
  </script>
  </body>
</html>
