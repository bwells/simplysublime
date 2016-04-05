if !has('python')
	finish
endif

let s:script_dir = escape(expand('<sfile>:p:h'), '\')

execute 'pyfile ' . s:script_dir . '/simplysublime.py'

function! BWFuzzyMatch(items, str, limit, mmode, ispath, crfile, regex)
	call clearmatches()
	if a:str == ''
		return a:items[0:a:limit]
	endif

	py fuzzy_matches()
	return s:results
endfunction
