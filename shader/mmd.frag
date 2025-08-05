#version 330 core
out vec4 FragColor;

in vec3 vsPos;
in vec3 vsNor;
in vec2 vsUV;

uniform vec3 uDiffuse;
uniform float uAlpha;
uniform vec3 uSpecular;
uniform float uSpecularPower;
uniform vec3 uAmbient;
uniform vec3 uLightColor;
uniform vec3 uLightPos;
uniform vec3 uViewPos;

uniform int uTexMode;
uniform sampler2D uTex;

uniform int uSphereTexMode;
uniform sampler2D uSphereTex;

uniform int uToonTexMode;
uniform sampler2D uToonTex;

void main() {
	vec3 eyeDir = normalize(uViewPos-vsPos);
	vec3 lightDir = normalize(uLightPos-vsPos);
	vec3 nor = normalize(vsNor);
	float ln = dot(nor, lightDir);
	ln = clamp(ln + 0.5, 0.0, 1.0);
	vec3 color = vec3(0.0, 0.0, 0.0);
	float alpha = uAlpha;
	vec3 diffuseColor = uDiffuse * uLightColor;
	color = diffuseColor;
	color += uAmbient;
	color = clamp(color, 0.0, 1.0);

	// 计算普通贴图
    if (uTexMode != 0) {
		vec4 texColor = texture(uTex, vsUV);
        color *= texColor.rgb;
		alpha *= texColor.a;
    }

	// 计算高光贴图
	if (uSphereTexMode != 0) {
		vec2 spUV = vec2(0.0);
		spUV.x = nor.x * 0.5 + 0.5;
		spUV.y = 1.0 - (nor.y * 0.5 + 0.5);
		vec3 spColor = texture(uSphereTex, spUV).rgb;
		if (uSphereTexMode == 1) { // 乘法
			color *= spColor;
		} else if (uSphereTexMode == 2) { // 加发
			color += spColor;
		} // 其他情况不支持
	}

	// 计算卡通映射
	if (uToonTexMode != 0) {
		vec3 toonColor = texture(uToonTex, vec2(0.0, ln)).rgb;
		color *= toonColor;
	}

	// 计算高光
	vec3 specular = vec3(0.0);
	if (uSpecularPower > 0) {
		vec3 halfVec = normalize(eyeDir + lightDir);
		vec3 specularColor = uSpecular * uLightColor;
		specular += pow(max(0.0, dot(halfVec, nor)), uSpecularPower) * specularColor;
	}

	// 添加高光返回最终结果
	color += specular;
	FragColor = vec4(color, alpha);
}
