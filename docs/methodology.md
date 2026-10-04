# Methodology: openphysics-steady-state-v1

This is a simultaneous design snapshot with immediate positive load contributions.
It is not Manual J or a dynamic building simulation. Heat flow is W, area m²,
temperature °C (differences K), pressure Pa, airflow m³/s, and humidity ratio kg
water/kg dry air. Model version and result IDs are included in every calculation.

## Conduction and assemblies

`Q = U × A_net × ΔT`, from the definition of thermal transmittance and steady-state
one-dimensional heat transfer. Heating uses `max(0, T_indoor - T_adjacent)`;
cooling uses `max(0, T_adjacent - T_indoor)`. Negative component credits are not
subtracted. This conservative clipping policy is explicit in each equation and
does not describe a full signed building heat balance.

`A_net = A_gross - sum(opening areas)`. Wall, ceiling and floor surfaces follow
the same boundary logic; windows and doors have their own assembly conduction.
An opening references a parent surface in the same room. An omitted opening
adjacency inherits the parent boundary. Opening area cannot exceed gross area.

Assemblies specify exactly one of direct U or explicit series R-values:
`U = 1 / sum(R_layers)`. No film resistance, framing fraction or material lookup
is silently added. Parallel paths require a defensible externally derived effective
U/R supplied with provenance. Each layer is retained in the calculation inputs.

Outdoor boundaries use project design temperatures. Conditioned boundaries have
zero temperature difference. Attic, garage, crawlspace, basement and adjacent-unit
boundaries require explicit seasonal adjacent temperatures. Ground/slab boundaries
are errors until a defensible ground-coupled method is implemented.

Opaque solar absorption, radiant exchange, thermal storage and time lag are
unsupported. Outdoor dry-bulb conduction alone cannot capture sun-heated opaque
surfaces. Apply results only with those limitations understood.

## Airflow

Natural ACH: `Vdot = ACH × building volume / 3600`.
Explicit airflow: use the supplied m³/s.
ACH50: `ACH_natural = ACH50 / supplied_conversion_factor`; then the natural-ACH
equation. The divisor is a user-selected assumption, never an implicit constant
or a claim of a design infiltration wind/stack model.

Infiltration is distributed in proportion to room volume. Ventilation is either
explicit room airflows summing to the supplied building airflow (absolute tolerance
1e-9 m³/s), or volume-proportional allocation with a visible assumption. Infiltration
and ventilation remain separate. Rates are interpreted at indoor design conditions.
There is no ventilation heat recovery, exhaust-induced infiltration interaction,
or airtightness/pressure-network calculation; entering overlapping rates can
double-count outdoor air.

## Psychrometrics and outdoor-air heat transfer

Saturation pressure uses PsychroLib 2.5.0's SI ice/water correlations and triple-point
branch, valid -100..200 °C. Humidity from RH is
`W = 0.621945 × RH × p_sat / (p - RH × p_sat)`.
Wet bulb uses the separate above/below-freezing PsychroLib energy-balance equations.
Require wet bulb ≤ dry bulb, positive pressure greater than vapor pressure, and
physically valid humidity. Exactly one outdoor RH or wet bulb is required. Values
are not silently clamped to a minimum humidity ratio.

Dry-air density in moist air:
`rho_dry = p / [287.042 × (T_indoor + 273.15) × (1 + 1.607858 × W_indoor)]`.
The constants and equations are adapted from the MIT-licensed PsychroLib SI
specific-volume and enthalpy functions; see the pinned provenance and notice.

Heating uses dry-air density (`W = 0`) at the indoor heating temperature and
`Q = rho_dry × Vdot × 1006 × max(0, T_indoor - T_outdoor)`.
The dry-air heating assumption is recorded in each dependent node.

Cooling sensible:
`Q_s = rho_dry × Vdot × (1006 + 1860 × W_indoor) × max(0, T_outdoor - T_indoor)`.
Cooling latent:
`Q_l = rho_dry × Vdot × 2501000 × max(0, W_outdoor - W_indoor)`.
2501000 J/kg is the zero-°C water-vapor enthalpy intercept in the adopted
approximation, not a temperature-dependent latent heat model. This component
split does not exactly equal the full moist-air enthalpy difference; it omits the
water-vapor sensible cross term associated with differing humidity ratios.
Dryer outdoor air is not given a negative latent-load credit.

These functions are encapsulated in `psychrometrics`, validated against independently
executed PsychroLib reference cases, and used by both outdoor-air components.

## Glazing solar and internal gains

`Q_solar = A × I_plane × SHGC × shading_factor`.
The definition of SHGC is the ratio of solar heat gain through glazing to incident
solar radiation. Plane irradiance includes whatever direct/diffuse ground-reflected
components the user has chosen. SHGC/shading are fractions in [0,1]. An explicit
irradiance is required, even when zero. No orientation-to-irradiance algorithm or
shading geometry inference is performed. All windows must represent a common
design snapshot; independent per-orientation peak values would overstate a
simultaneous block load.

The engine treats transmitted/admitted instantaneous solar gain as immediate
sensible cooling load. It does not apply response factors, cooling-load factors,
mass/storage delay, or frame-specific optical models. Nonzero solar nodes warn
about this limitation. Non-outdoor glazing must specify zero solar irradiance.

Occupants: `Q = count × explicitly supplied W/person`, separately sensible/latent.
Lighting and other sensible/latent gains are explicit W inputs. No occupant
activity, appliance, lighting diversity or schedule values are defaulted. Users
must supply gains for the intended common design snapshot and label assumptions.

## Aggregation and confidence

Room component sums form zone sums and building sums. Cooling total is sensible
plus latent. Heating receives no offsetting internal/solar credits. There are no
safety factors, arbitrary adjustments, equipment capacities or duct penalties.

Input evidence can be measured, user-entered, derived, imported, dataset-based,
public-reference-based, or a labeled default. Omitted evidence classifies values
as supplied project input; it does not introduce a numeric default. Assumptions
on inputs are carried into leaf and aggregate nodes. This is qualitative tracking,
not a probability distribution or statistical confidence interval.

