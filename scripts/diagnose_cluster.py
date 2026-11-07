import urllib.request
import json

def check_prometheus():
    try:
        url = 'http://192.168.220.20:9090/api/v1/targets'
        with urllib.request.urlopen(url, timeout=4) as resp:
            data = json.loads(resp.read().decode('utf-8'))
            targets = data.get('data', {}).get('activeTargets', [])
            print("=== PROMETHEUS ACTIVE TARGETS ===")
            for t in targets:
                pool = t.get('scrapePool', '')
                health = t.get('health', '')
                err = t.get('lastError', '')
                print(f"Pool: {pool:15} | Health: {health:5} | Error: {err}")
    except Exception as e:
        print("Error checking Prometheus:", e)

def check_sql_exporter():
    try:
        url = 'http://192.168.220.20:9399/metrics'
        with urllib.request.urlopen(url, timeout=4) as resp:
            print("sql-exporter (9399) is UP!")
            lines = resp.read().decode('utf-8', errors='ignore').split('\n')
            print(f"Total metric lines returned: {len(lines)}")
            sql_lines = [l for l in lines if 'sqlserver' in l and not l.startswith('#')][:10]
            for s in sql_lines:
                print("  Sample metric:", s)
    except Exception as e:
        print("sql-exporter (9399) is DOWN or unreachable:", e)

if __name__ == '__main__':
    check_sql_exporter()
    print()
    check_prometheus()

