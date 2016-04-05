import vim
from .. import match_all

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
