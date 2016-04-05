import time

import fuzzy

def load_corpus(filename):

    with open(filename) as f:
        lines = f.readlines()

    return lines

def bench(pattern, runs=1000):

    corpus = load_corpus('files')
    times = []
    for i in range(runs):
        start = time.time()
        results = fuzzy.match_all(corpus, pattern)
        end = time.time()
        duration = end - start
        times.append(duration)

    avg = sum(n for n in times) / runs
    avg_ms = avg * 1000

    print '"%s" took avg. %0.2fms across %s runs' % (pattern, avg_ms, runs)

def bench_format(runs=1000):

    times = []

    # cont test acc -> /erp/controllers/testing/accounts.py
    indices = [4,5,6,7,16,17,18,19,24,25,26]
    string = "/erp/controllers/testing/accounts.py"

    for i in range(runs):
        start = time.time()
        result = fuzzy.format_match(indices, string)
        end = time.time()
        duration = end - start
        times.append(duration)

    avg = sum(n for n in times) / runs
    avg_ms = avg * 1000

    print 'took avg. %0.2fms across %s runs' % (avg_ms, runs)

if __name__ == '__main__':
    bench('cont test acc', runs=100)
    # bench_format(runs=10000)
