clean:
	@rm -rf example
	@go clean

# Tangle out the example in the README.org using Emacs org-babel
example:
	@docker run -it --rm -v $$(pwd):/gastly silex/emacs bash -c "cd /gastly ; install -o $$(stat -c '%u' ./README.org) -g $$(stat -c '%g' ./README.org) -d -m 755 /gastly/example ; emacs --batch -l org --eval '(org-babel-tangle-file \"/gastly/README.org\")'"
	@echo "See the example implementation in the 'example' directory!"

test:
	@go test -v
