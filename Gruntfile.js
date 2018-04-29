var yaml = require("js-yaml");
var S = require("string");

var CONTENT_PATH_PREFIX = "content";

module.exports = function(grunt) {

    grunt.registerTask("lunr-index", function() {

        grunt.log.writeln("Build pages index");

        var indexPages = function() {
            var pagesIndex = [];
            grunt.file.recurse(CONTENT_PATH_PREFIX, function(abspath, rootdir, subdir, filename) {
                // grunt.verbose.writeln("Parse file:",abspath);
                // grunt.log.writeln("Parse file:", abspath);
                pagesIndex.push(processFile(abspath, filename));
            });

            return pagesIndex;
        };

        var processFile = function(abspath, filename) {
            var pageIndex;

            if (S(filename).endsWith(".html")) {
                pageIndex = processHTMLFile(abspath, filename);
            } else {
                pageIndex = processMDFile(abspath, filename);
            }

            return pageIndex;
        };

        var processHTMLFile = function(abspath, filename) {
            var content = grunt.file.read(abspath);
            var pageName = S(filename).chompRight(".html").s;
            var href = S(abspath)
                .chompLeft(CONTENT_PATH_PREFIX).s;
            return {
                title: pageName,
                href: href,
                content: S(content).trim().stripTags().stripPunctuation().s
            };
        };

        var processMDFile = function(abspath, filename) {
            var content = grunt.file.read(abspath);
            var pageIndex;

            // separate the Front Matter from the content and parse it
            content = content.split("---");
            // grunt.log.writeln(content[1]);

            var frontMatter;
            try {
                frontMatter = yaml.load(content[1]);
            } catch (e) {
                grunt.log.writeln(e.message);
            }

            // grunt.log.writeln(JSON.stringify(frontMatter));

            var href = S(abspath).chompLeft(CONTENT_PATH_PREFIX).chompRight(".md").s;

            if (filename === ".DS_Store") {
              return
            }

            if (filename === "_index.md") {
                href = "/"
                // S(abspath).chompLeft(CONTENT_PATH_PREFIX).chompRight(filename).s;
            }
            var m = abspath.match(/^content\/page\/(.+)\.md/);
            if (m != null) {
              href = "/" + m[1]
            }

            // Build Lunr index for this page
            pageIndex = {
                title: frontMatter.title,
                tags: frontMatter.tags,
                href: href.toLowerCase(),
                content: S(content[2]).stripTags().stripPunctuation().s
            };

            grunt.log.writeln(JSON.stringify(pageIndex));
            return pageIndex;
        };

        grunt.file.write("static/js/lunr/index.json", JSON.stringify(indexPages()));
        grunt.log.ok("Index built");
    });
};
