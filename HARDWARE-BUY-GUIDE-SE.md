# Homelab Buy Guide (Linux-capable) — 4 good options

> Goal: run Cubby + more services, stable storage, long-term growth.  
> Store links are Swedish (primarily Dustin), with direct product/search links.

## 1) Raspberry Pi option (best SBC route)
**Recommended to buy**
- **Raspberry Pi 5 Model B 8GB** (host)
- **Official PSU + active cooling/case**
- **NVMe storage setup** (preferred) or at least high-endurance SD for boot + external SSD for data

**Links**
- Raspberry Pi 5 Model B 8GB (Dustin):  
  https://www.dustin.se/product/5011369899/5-model-b-8gb
- Raspberry Pi 5 8GB Kit (Dustin):  
  https://www.dustin.se/product/5020049324/dnc-raspberry-pi-5-8gb-kit

**Notes**
- Run Linux (Debian/Ubuntu Server 64-bit)
- For database stability: keep SQLite/data on SSD, not only SD-card.

---

## 2) AMD Ryzen mini PC (best Linux value/performance)
**Recommended to buy**
- **Mini PC with Ryzen 7/AI-class CPU**, **32GB RAM minimum** (prefer 64GB if many services)
- **1TB NVMe minimum** (prefer 2TB)

**Links**
- Dustin Ryzen mini PC search/results:  
  https://www.dustin.se/search/ryzen%20mini%20pc
- Example ASUS ExpertCenter PN54 (Dustin):  
  https://www.dustin.se/product/5020064959/expertcenter-pn54
- Example Lenovo ThinkCentre M75q G5 Tiny Ryzen 7 PRO (Dustin):  
  https://www.dustin.se/product/5020023425/thinkcentre-m75q-g5-tiny

**Notes**
- Great for Docker + many services on Linux.
- Prefer models where RAM/SSD are upgradeable.

---

## 3) Intel NUC-class (compact and very Linux-friendly)
**Recommended to buy**
- **Intel Core Ultra 7 or better**
- **32GB RAM minimum**, preferably 64GB for many containers
- **1–2TB NVMe SSD**

**Links**
- Dustin Intel NUC search/results:  
  https://www.dustin.se/search/intel%20nuc
- Example ASUS NUC 15 Pro U7 265H Barebone (Dustin):  
  https://www.dustin.se/product/5020064039/nuc-15-pro-u7-265h-vpro-tall-barebone
- Example ASUS NUC 14 Pro+ Ultra 9 (Dustin):  
  https://www.dustin.se/product/5020070857/nuc-14-pro-rnuc14rvsu9089a2i-intel-core-ultra-9-185h-32-gb-ddr5-sdram-1-tb-ssd-windows-11-home-ucff-mini-pc-vit

**Notes**
- Barebone units are great if you want your own Linux-first RAM/SSD choices.

---

## 4) Mac mini (powerful hardware, but Linux caveat)
**Recommended to buy (if you still want Mac mini hardware)**
- **Mac mini M4 / M4 Pro**, 24GB+ RAM if running many services
- External high-quality SSD for backups/data if needed

**Links**
- Dustin Mac mini search/results:  
  https://www.dustin.se/search/mac%20mini
- Example Mac mini (2024) M4 32GB 1TB (Dustin):  
  https://www.dustin.se/product/5020028787/mac-mini-2024
- Example Mac mini (2024) M4 Pro 24GB 1TB (Dustin):  
  https://www.dustin.se/product/5020028796/mac-mini-2024

**Important Linux note**
- Mac mini is excellent hardware, but if your requirement is **Linux as host OS**, AMD/Intel mini PCs are easier and cleaner for full native Linux operations.

---

## What I’d buy (if “no budget” + “Linux”)
1. **AMD Ryzen mini PC, 64GB RAM, 2TB NVMe** (primary)
2. Or **Intel NUC-class, 64GB RAM, 2TB NVMe**
3. Add **UPS** and automated backups from day one

This gives you the best long-term Linux homelab platform for Cubby + additional services.
