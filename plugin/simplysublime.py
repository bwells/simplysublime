import vim

# from CPSM plugin
def _escape_and_quote(s):
    return '"' + s.replace("\\", "\\\\").replace('"', '\\"') + '"'

def fuzzy_matches():
    try:
        items = vim.eval('a:items')
        astr = vim.eval('a:str')
        limit = int(vim.eval('a:limit'))

        results = match_all(items, astr, limit=limit)

        results = [r[1] for r in results]

        results = ",".join(map(_escape_and_quote, results))
        vim.command("let s:results = [%s]" % results)

        # for r in results:
            # print r
    except Exception as ex:
        vim.command("let s:results = [%s]" % _escape_and_quote(
                            "ERROR: " + str(ex)
            ))

import itertools
import heapq

def get_names():
    with open('files') as f:
        lines = f.readlines()
        lines = [l.strip().strip(',').strip('"') for l in lines]

    return lines

def match_all(items, pattern, limit=10):

    results = []
    for item in items:
        results.append(fuzzy(pattern, item))

    results = heapq.nlargest(limit, results)

    formatted_results = [format_match(r[1], r[2]) for r in results]

    return formatted_results

def format_match(indices, string, fmt='shell'):

    if fmt == 'shell':
        s = "\033[31;1m" # red; bold
        e = "\033[0m"

    else:
        s = '<b>'
        e = '</b>'

    formatted_str = []
    last_idx = 0
    for idx in indices:
        formatted_str.append('%s%s%s%s' % (string[last_idx:idx], s, string[idx], e))
        last_idx = idx + 1
    formatted_str.append(string[last_idx:])
    formatted_str = ''.join(formatted_str)

    return formatted_str

def fuzzy(pattern, string):

    LAST_WAS_SEPARATOR_BONUS = 5
    LAST_WAS_PATH_SEPARATOR_BONUS = 8
    MATCHED_CHAR_BONUS = 2
    SEPARATOR_MATCH_BONUS = 1
    GAP_PENELTY = 2

    score = 0
    matched_indices = []
    last_str_was_sep = True # start True to match first char in string
    sep_matches = 0
    last_seps = 0
    path_last_seps = 0

    pattern_idx = 0
    str_idx = 0

    SEPARATORS = ('_', ' ', '/')
    PATH_SEPARATORS = ('/', '\\')

    str_len = len(string)
    pattern_len = len(pattern)

    for str_idx in range(str_len):
        pattern_char = pattern[pattern_idx] if pattern_idx < pattern_len else None
        str_char = string[str_idx]

        # pattern_is_sep = pattern_char in SEPARATORS
        str_is_sep = str_char in SEPARATORS
        # print 'checking %s against %s [%s, %s]' % (pattern_char, str_char, pattern_idx, str_idx)
        if pattern_char and pattern_char.lower() == str_char.lower():
            # print '  matched %s' % str_char
            matched_indices.append(str_idx)
            score += MATCHED_CHAR_BONUS
            pattern_idx += 1
            if last_str_was_sep:
                # print 'Last sep: "%s"' % last_str_char
                if last_str_char in PATH_SEPARATORS:
                    score += LAST_WAS_PATH_SEPARATOR_BONUS
                    path_last_seps += 1
                else:
                    score += LAST_WAS_SEPARATOR_BONUS
                    last_seps += 1

        # elif str_is_sep and pattern_is_sep:
        elif str_is_sep and pattern_char in SEPARATORS:
        # print '  matched SEP'
            score += SEPARATOR_MATCH_BONUS
            sep_matches += 1
            pattern_idx += 1

        last_str_was_sep = str_is_sep
        last_str_char = str_char

    run_length = 0
    gaps = 0
    # runs = []
    for last_idx, idx in pairwise(matched_indices):
        if idx - last_idx == 1:
            run_length += 1
        else:
            # runs.append(run_length+1)
            score += (run_length + 1) ** 2 # + 1 because fence posts
            run_length = 0
            gaps += 1

    if run_length > 1:
        score += (run_length + 1) ** 2
        # runs.append(run_length+1)

    score -= gaps * GAP_PENELTY

    # score -= len(string) - len(matched_indices)

    # fmt = format_match(matched_indices, string)

    SCORE_REPORT = False
    if SCORE_REPORT:
        print "Matched Chars:\t%02d" % ( len(matched_indices) * MATCHED_CHAR_BONUS
                )
        print "Sep Matches:\t%02d" % (sep_matches * SEPARATOR_MATCH_BONUS)
        print "Last Seps:\t%02d" % (last_seps * LAST_WAS_SEPARATOR_BONUS)
        print "Path Last Seps:\t%02d" % (path_last_seps * LAST_WAS_PATH_SEPARATOR_BONUS)
        print "Runs:\t\t%02d" % sum(r ** 2 for r in runs)
        print "- Gaps:\t\t%02d" % (gaps * 2)
        # print "- Len:\t\t%02d" % (len(string) - len(matched_indices))
        print score, fmt, '\n'

    # return score, fmt
    return score, matched_indices, string

def pairwise(iterable):
    """ Creates an iterator that iterates the original in (n, n+1) pairs
        i -> (i0, i1), (i1, i2), (i2, i3), ... (iN-1, iN) """
    return itertools.izip(iterable, itertools.islice(iterable, 1, None))


