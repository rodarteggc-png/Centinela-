#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
CENTINELA OBSERVABILITY — Glassmorphism & Centinela Rebranding Engine
Applies:
1. Centinela rebranding (titles, headers, row descriptions).
2. Clean Glassmorphism styling with <img> logo (no raw SVG leaks).
3. Fallback variable presets for SQL-CLUSTER-02.
"""

import json
from pathlib import Path

dashboards_dir = Path(__file__).resolve().parent.parent / "grafana" / "dashboards"

# -----------------------------------------------------------------------------
# 1. Cockpit Header HTML for Fleet Global View
# -----------------------------------------------------------------------------
global_view_header = """<div style="background: rgba(13, 20, 36, 0.75); backdrop-filter: blur(20px) saturate(180%); -webkit-backdrop-filter: blur(20px) saturate(180%); border: 1px solid rgba(255, 255, 255, 0.1); border-top: 2px solid #00F2FE; border-radius: 14px; padding: 18px 24px; box-shadow: 0 12px 40px rgba(0,0,0,0.5), inset 0 1px 0 rgba(255,255,255,0.1);">
  <div style="display: flex; justify-content: space-between; align-items: center; flex-wrap: wrap; gap: 16px;">
    <div style="display: flex; align-items: center; gap: 18px;">
      <div style="background: linear-gradient(135deg, rgba(0,242,254,0.2) 0%, rgba(0,82,212,0.2) 100%); border: 1px solid rgba(0,242,254,0.4); width: 48px; height: 48px; border-radius: 12px; display: flex; align-items: center; justify-content: center; box-shadow: 0 0 20px rgba(0,242,254,0.35);">
        <img src="public/img/grafana_icon.svg" width="34" height="34" style="display: block; filter: drop-shadow(0 0 8px rgba(0,242,254,0.6));" alt="Centinela" />
      </div>
      <div>
        <div style="display: flex; align-items: center; gap: 10px; flex-wrap: wrap;">
          <span style="font-size: 22px; font-weight: 800; color: #f8fafc; letter-spacing: 0.8px;">CENTINELA OBSERVABILITY</span>
          <span style="background: linear-gradient(135deg, rgba(6,182,212,0.25) 0%, rgba(59,130,246,0.25) 100%); color: #38bdf8; border: 1px solid rgba(56,189,248,0.4); font-size: 11px; font-weight: 800; padding: 3px 12px; border-radius: 9999px; text-transform: uppercase;">Fleet Global View</span>
          <span style="background: rgba(16,185,129,0.15); color: #10B981; border: 1px solid rgba(16,185,129,0.3); font-size: 11px; font-weight: 700; padding: 3px 10px; border-radius: 9999px;">Cluster Always On Activo</span>
        </div>
        <div style="font-size: 12px; color: #94a3b8; margin-top: 5px; display: flex; gap: 20px; flex-wrap: wrap;">
          <span><b>Cluster Activo:</b> SQL-CLUSTER-02</span>
          <span><b>Nodos Reales:</b> SQL-CLUSTER-02nodo1 &bull; SQL-CLUSTER-02nodo2</span>
          <span><b>Servidor Central:</b> Rocky Linux 9 (sda2110022 / 192.168.220.20)</span>
        </div>
      </div>
    </div>
    <div style="display: flex; gap: 12px; text-align: center;">
      <div style="background: rgba(255,255,255,0.03); backdrop-filter: blur(10px); border: 1px solid rgba(255,255,255,0.08); padding: 8px 16px; border-radius: 10px; box-shadow: 0 4px 15px rgba(0,0,0,0.2);">
        <div style="font-size: 10px; color: #64748b; font-weight: 800;">NODOS CLUSTER</div>
        <div style="font-size: 16px; font-weight: 900; color: #38bdf8;">2 Nodos (5000)</div>
      </div>
      <div style="background: rgba(255,255,255,0.03); backdrop-filter: blur(10px); border: 1px solid rgba(255,255,255,0.08); padding: 8px 16px; border-radius: 10px; box-shadow: 0 4px 15px rgba(0,0,0,0.2);">
        <div style="font-size: 10px; color: #64748b; font-weight: 800;">ESTADO LICENCIAS</div>
        <div style="font-size: 16px; font-weight: 900; color: #10B981;">100% LIBRE $0</div>
      </div>
    </div>
  </div>
  <div style="display: flex; gap: 10px; margin-top: 14px; border-top: 1px solid rgba(255,255,255,0.08); padding-top: 12px; flex-wrap: wrap;">
    <a href="/d/centinela-databases-global-view" style="text-decoration: none; background: linear-gradient(135deg, rgba(6,182,212,0.3) 0%, rgba(59,130,246,0.3) 100%); color: #ffffff; border: 1px solid rgba(56,189,248,0.5); font-weight: 700; font-size: 11px; padding: 6px 16px; border-radius: 8px; display: flex; align-items: center; gap: 8px; box-shadow: 0 0 12px rgba(56,189,248,0.2);">🌐 Centinela Global View</a>
    <a href="/d/centinela-sqlserver-cluster" style="text-decoration: none; background: rgba(255,255,255,0.04); color: #94a3b8; border: 1px solid rgba(255,255,255,0.08); font-weight: 700; font-size: 11px; padding: 6px 16px; border-radius: 8px; display: flex; align-items: center; gap: 8px;">🔬 Performance Investigator (PI)</a>
    <a href="#inventory-grid" style="text-decoration: none; background: rgba(255,255,255,0.04); color: #94a3b8; border: 1px solid rgba(255,255,255,0.08); font-weight: 700; font-size: 11px; padding: 6px 16px; border-radius: 8px; display: flex; align-items: center; gap: 8px;">📋 Inventario y Semaforos</a>
    <a href="#fleet-trend" style="text-decoration: none; background: rgba(255,255,255,0.04); color: #94a3b8; border: 1px solid rgba(255,255,255,0.08); font-weight: 700; font-size: 11px; padding: 6px 16px; border-radius: 8px; display: flex; align-items: center; gap: 8px;">📈 Tendencia de Carga AAS</a>
  </div>
</div>"""

# -----------------------------------------------------------------------------
# 2. Cockpit Header HTML for SQL Server Performance Investigator
# -----------------------------------------------------------------------------
pi_header = """<div style="background: rgba(13, 20, 36, 0.75); backdrop-filter: blur(20px) saturate(180%); -webkit-backdrop-filter: blur(20px) saturate(180%); border: 1px solid rgba(255, 255, 255, 0.1); border-top: 2px solid #00F2FE; border-radius: 14px; padding: 18px 24px; box-shadow: 0 12px 40px rgba(0,0,0,0.5), inset 0 1px 0 rgba(255,255,255,0.1);">
  <div style="display: flex; justify-content: space-between; align-items: center; flex-wrap: wrap; gap: 16px;">
    <div style="display: flex; align-items: center; gap: 18px;">
      <div style="background: linear-gradient(135deg, rgba(0,242,254,0.2) 0%, rgba(0,82,212,0.2) 100%); border: 1px solid rgba(0,242,254,0.4); width: 48px; height: 48px; border-radius: 12px; display: flex; align-items: center; justify-content: center; box-shadow: 0 0 20px rgba(0,242,254,0.35);">
        <img src="public/img/grafana_icon.svg" width="34" height="34" style="display: block; filter: drop-shadow(0 0 8px rgba(0,242,254,0.6));" alt="Centinela" />
      </div>
      <div>
        <div style="display: flex; align-items: center; gap: 10px; flex-wrap: wrap;">
          <span style="font-size: 22px; font-weight: 800; color: #f8fafc; letter-spacing: 0.8px;">CENTINELA OBSERVABILITY</span>
          <span style="background: linear-gradient(135deg, rgba(6,182,212,0.25) 0%, rgba(59,130,246,0.25) 100%); color: #38bdf8; border: 1px solid rgba(56,189,248,0.4); font-size: 11px; font-weight: 800; padding: 3px 12px; border-radius: 9999px; text-transform: uppercase;">Performance Investigator</span>
          <span style="background: rgba(16,185,129,0.15); color: #10B981; border: 1px solid rgba(16,185,129,0.3); font-size: 11px; font-weight: 700; padding: 3px 10px; border-radius: 9999px;">Cluster Always On</span>
        </div>
        <div style="font-size: 12px; color: #94a3b8; margin-top: 5px; display: flex; gap: 20px; flex-wrap: wrap;">
          <span><b>Cluster:</b> $cluster</span>
          <span><b>Instancia:</b> $target</span>
          <span><b>HADR:</b> Always On Availability Groups</span>
          <span><b>Base de Datos:</b> $database</span>
        </div>
      </div>
    </div>
    <div style="display: flex; gap: 12px; text-align: center;">
      <div style="background: rgba(255,255,255,0.03); backdrop-filter: blur(10px); border: 1px solid rgba(255,255,255,0.08); padding: 8px 16px; border-radius: 10px; box-shadow: 0 4px 15px rgba(0,0,0,0.2);">
        <div style="font-size: 10px; color: #64748b; font-weight: 800;">BUFFER CACHE</div>
        <div style="font-size: 16px; font-weight: 900; color: #10B981;">> 99.5%</div>
      </div>
      <div style="background: rgba(255,255,255,0.03); backdrop-filter: blur(10px); border: 1px solid rgba(255,255,255,0.08); padding: 8px 16px; border-radius: 10px; box-shadow: 0 4px 15px rgba(0,0,0,0.2);">
        <div style="font-size: 10px; color: #64748b; font-weight: 800;">PAGE LIFE (PLE)</div>
        <div style="font-size: 16px; font-weight: 900; color: #10B981;">> 3,000s</div>
      </div>
      <div style="background: rgba(255,255,255,0.03); backdrop-filter: blur(10px); border: 1px solid rgba(255,255,255,0.08); padding: 8px 16px; border-radius: 10px; box-shadow: 0 4px 15px rgba(0,0,0,0.2);">
        <div style="font-size: 10px; color: #64748b; font-weight: 800;">CORES FISICOS</div>
        <div style="font-size: 16px; font-weight: 900; color: #38bdf8;">16 vCPUs</div>
      </div>
    </div>
  </div>
  <div style="display: flex; gap: 10px; margin-top: 14px; border-top: 1px solid rgba(255,255,255,0.08); padding-top: 12px; flex-wrap: wrap;">
    <a href="/d/centinela-databases-global-view" style="text-decoration: none; background: rgba(255,255,255,0.04); color: #94a3b8; border: 1px solid rgba(255,255,255,0.08); font-weight: 700; font-size: 11px; padding: 6px 16px; border-radius: 8px; display: flex; align-items: center; gap: 8px;">🌐 Centinela Global View</a>
    <a href="/d/centinela-sqlserver-cluster?var-cluster=$cluster&var-target=$target" style="text-decoration: none; background: linear-gradient(135deg, rgba(6,182,212,0.3) 0%, rgba(59,130,246,0.3) 100%); color: #ffffff; border: 1px solid rgba(56,189,248,0.5); font-weight: 700; font-size: 11px; padding: 6px 16px; border-radius: 8px; display: flex; align-items: center; gap: 8px; box-shadow: 0 0 12px rgba(56,189,248,0.2);">🔬 Performance Investigator (PI)</a>
    <a href="#top-sql" style="text-decoration: none; background: rgba(255,255,255,0.04); color: #94a3b8; border: 1px solid rgba(255,255,255,0.08); font-weight: 700; font-size: 11px; padding: 6px 16px; border-radius: 8px; display: flex; align-items: center; gap: 8px;">📋 Top SQL Statements</a>
    <a href="#advisories" style="text-decoration: none; background: rgba(255,255,255,0.04); color: #38bdf8; border: 1px solid rgba(56,189,248,0.2); font-weight: 700; font-size: 11px; padding: 6px 16px; border-radius: 8px; display: flex; align-items: center; gap: 8px;">💡 Centinela Advisories</a>
    <a href="#storage" style="text-decoration: none; background: rgba(255,255,255,0.04); color: #94a3b8; border: 1px solid rgba(255,255,255,0.08); font-weight: 700; font-size: 11px; padding: 6px 16px; border-radius: 8px; display: flex; align-items: center; gap: 8px;">💾 Storage Runway</a>
  </div>
</div>"""

# -----------------------------------------------------------------------------
# 3. Add $database variable to centinela-sqlserver-cluster.json
# -----------------------------------------------------------------------------
def fix_sqlserver_variables(data):
    templating = data.setdefault("templating", {}).setdefault("list", [])
    
    # Check if 'cluster' has fallback values
    for v in templating:
        if v.get("name") == "cluster":
            v["current"] = {"selected": True, "text": "SQL-CLUSTER-02", "value": "SQL-CLUSTER-02"}
            v["options"] = [{"selected": True, "text": "SQL-CLUSTER-02", "value": "SQL-CLUSTER-02"}]
            v["query"] = {"query": "label_values(sqlserver_connections_active, cluster)", "refId": "StandardVariableQuery"}
            v["definition"] = "label_values(sqlserver_connections_active, cluster)"
        if v.get("name") == "target":
            v["current"] = {"selected": True, "text": "All", "value": "$__all"}
            v["options"] = [
                {"selected": True, "text": "All", "value": "$__all"},
                {"selected": False, "text": "SQL-CLUSTER-02nodo1", "value": "SQL-CLUSTER-02nodo1"},
                {"selected": False, "text": "SQL-CLUSTER-02nodo2", "value": "SQL-CLUSTER-02nodo2"}
            ]
            v["query"] = {"query": "label_values(sqlserver_connections_active{cluster=~\"$cluster\"}, target)", "refId": "StandardVariableQuery"}
            v["definition"] = "label_values(sqlserver_connections_active{cluster=~\"$cluster\"}, target)"
    
    # Check if 'database' variable exists
    has_db = any(v.get("name") == "database" for v in templating)
    if not has_db:
        templating.append({
            "current": {"selected": True, "text": "All", "value": "$__all"},
            "datasource": {"type": "prometheus", "uid": "centinela-prometheus"},
            "definition": "label_values(sqlserver_connections_active{cluster=~\"$cluster\"}, database_name)",
            "hide": 0,
            "includeAll": True,
            "label": "Base de Datos",
            "multi": False,
            "name": "database",
            "options": [{"selected": True, "text": "All", "value": "$__all"}],
            "query": {"query": "label_values(sqlserver_connections_active{cluster=~\"$cluster\"}, database_name)", "refId": "StandardVariableQuery"},
            "refresh": 1,
            "regex": "",
            "skipUrlSync": False,
            "sort": 1,
            "type": "query"
        })

def process_file(file_path, title, header_html=None, is_sql_pi=False):
    if not file_path.exists():
        print(f"[!] File not found: {file_path}")
        return
    with open(file_path, "r", encoding="utf-8") as f:
        data = json.load(f)
    
    data["title"] = title
    
    if is_sql_pi:
        fix_sqlserver_variables(data)
    
    for p in data.get("panels", []):
        if p.get("type") != "row":
            p["transparent"] = True
        
        if p.get("id") == 99 and header_html:
            p.setdefault("options", {})["content"] = header_html
        
        # Row title updates
        if p.get("type") == "row":
            t = p.get("title", "")
            if "RESUMEN DE SALUD" in t or "FLEET HEALTH" in t:
                p["title"] = "🏛️ CENTINELA — SALUD GLOBAL DE LA FLOTA CORPORATIVA"
            elif "PERFORMANCE INVESTIGATOR" in t:
                p["title"] = "🔬 CENTINELA PERFORMANCE INVESTIGATOR (PI) — WORKLOAD & RESOURCE TIME (AAS)"
            elif "FOGLIGHT ADVISORY" in t or "ADVISORY" in t:
                p["title"] = "💡 CENTINELA ADVISORY ENGINE — ÍNDICES FALTANTES & OPTIMIZACIONES"
        
        # Panel descriptions & titles
        p_title = p.get("title", "")
        if "Foglight PI" in p_title:
            p["title"] = p_title.replace("Foglight PI", "Centinela PI")
        if "Quest Foglight" in p_title:
            p["title"] = p_title.replace("Quest Foglight", "Centinela")
        
        p_desc = p.get("description", "")
        if "Quest Foglight" in p_desc:
            p["description"] = p_desc.replace("Quest Foglight", "Centinela")

    with open(file_path, "w", encoding="utf-8") as f:
        json.dump(data, f, indent=2, ensure_ascii=False)
    print(f"[OK] Upgraded: {file_path.name}")

# Process all dashboards
process_file(dashboards_dir / "centinela-databases-global-view.json", "Centinela — Fleet Global View", global_view_header)
process_file(dashboards_dir / "centinela-sqlserver-cluster.json", "Centinela — SQL Server Performance Investigator", pi_header, is_sql_pi=True)
print("\n[OK] All dashboards upgraded successfully without SVG leak!")
