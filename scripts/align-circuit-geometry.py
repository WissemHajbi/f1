"""Align a curated OSM circuit centerline to one locally stored OpenF1 reference lap."""
import argparse, json, math, sqlite3
from pathlib import Path


def resample(points, count=250):
    if points[0] != points[-1]: points = points + [points[0]]
    distances = [0.0]
    for a, b in zip(points, points[1:]): distances.append(distances[-1] + math.hypot(b[0]-a[0], b[1]-a[1]))
    result, segment = [], 0
    for index in range(count):
        distance = distances[-1] * index / count
        while segment + 1 < len(distances) and distances[segment + 1] < distance: segment += 1
        ratio = (distance-distances[segment]) / (distances[segment+1]-distances[segment] or 1)
        a, b = points[segment], points[segment+1]
        result.append((a[0]+ratio*(b[0]-a[0]), a[1]+ratio*(b[1]-a[1])))
    return result


def fit(source, target):
    sc = (sum(x for x, _ in source)/len(source), sum(y for _, y in source)/len(source))
    tc = (sum(x for x, _ in target)/len(target), sum(y for _, y in target)/len(target))
    centered_source = [(x-sc[0], y-sc[1]) for x, y in source]
    centered_target = [(x-tc[0], y-tc[1]) for x, y in target]
    a = sum(x*u+y*v for (x, y), (u, v) in zip(centered_source, centered_target))
    b = sum(x*v-y*u for (x, y), (u, v) in zip(centered_source, centered_target))
    denominator = sum(x*x+y*y for x, y in centered_source)
    scale_cos, scale_sin = a/denominator, b/denominator
    transformed = [(tc[0]+scale_cos*x-scale_sin*y, tc[1]+scale_sin*x+scale_cos*y) for x, y in centered_source]
    error = math.sqrt(sum((x-u)**2+(y-v)**2 for (x, y), (u, v) in zip(transformed, target))/len(target))
    return error, (sc, tc, scale_cos, scale_sin)


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--input', required=True); parser.add_argument('--output', required=True)
    parser.add_argument('--database', default='data/oidysts.db'); parser.add_argument('--session', type=int, required=True)
    args = parser.parse_args()
    geojson = json.loads(Path(args.input).read_text(encoding='utf8')); feature = geojson['features'][0]
    coordinates = feature['geometry']['coordinates']; latitude = sum(p[1] for p in coordinates)/len(coordinates)
    origin = coordinates[0]
    osm = [((lon-origin[0])*111320*math.cos(math.radians(latitude)), (lat-origin[1])*110540) for lon, lat in coordinates]
    db = sqlite3.connect(args.database)
    reference = db.execute('''SELECT l.driver_number,l.lap_number,l.date_start,l.lap_duration,COUNT(*) samples
      FROM laps l JOIN location_samples p ON p.session_key=l.session_key AND p.driver_number=l.driver_number
       AND julianday(p.sampled_at)>=julianday(l.date_start)
       AND julianday(p.sampled_at)<=julianday(l.date_start)+l.lap_duration/86400.0
      WHERE l.session_key=? AND l.date_start IS NOT NULL AND l.lap_duration IS NOT NULL AND l.is_pit_out_lap=0
      GROUP BY l.driver_number,l.lap_number HAVING samples>=250 ORDER BY l.lap_duration ASC LIMIT 1''', (args.session,)).fetchone()
    target = db.execute('''SELECT x,y FROM location_samples WHERE session_key=? AND driver_number=?
      AND julianday(sampled_at)>=julianday(?) AND julianday(sampled_at)<=julianday(?)+?/86400.0 ORDER BY sampled_at''',
      (args.session, reference[0], reference[2], reference[2], reference[3])).fetchall()
    source_samples, target_samples = resample(osm), resample(target)
    best = None
    for reverse in (False, True):
        base = list(reversed(source_samples)) if reverse else source_samples
        for shift in range(len(base)):
            candidate = base[shift:]+base[:shift]; error, transform = fit(candidate, target_samples)
            if best is None or error < best[0]: best = (error, reverse, shift, transform)
    error, reverse, shift, (source_center, target_center, scale_cos, scale_sin) = best
    aligned = []
    for x, y in osm:
        x, y = x-source_center[0], y-source_center[1]
        aligned.append({'x': round(target_center[0]+scale_cos*x-scale_sin*y), 'y': round(target_center[1]+scale_sin*x+scale_cos*y)})
    output = {'session_key': args.session, 'points': aligned, 'estimated_width_m': feature['properties']['estimated_width_m'],
      'attribution': feature['properties']['attribution'], 'source_url': feature['properties']['source_url'],
      'geometry_accuracy': 'OSM centerline with estimated width; not official asphalt boundaries',
      'alignment': {'reference_driver': reference[0], 'reference_lap': reference[1], 'rms_error_m': round(error/math.hypot(scale_cos, scale_sin), 2), 'reversed': reverse, 'shift': shift}}
    Path(args.output).write_text(json.dumps(output, indent=2), encoding='utf8')
    print(f"wrote {len(aligned)} points; alignment RMS {output['alignment']['rms_error_m']} m")

if __name__ == '__main__': main()
