all: libz3.a

clean:
	rm -rf vendor
	rm -f libz3.a

libz3.a:
	cd vendor/z3 && python3 scripts/mk_make.py --staticlib
	cd vendor/z3/build && ${MAKE} -j10
	cp vendor/z3/build/libz3.a .

.PHONY: all clean libz3.a
