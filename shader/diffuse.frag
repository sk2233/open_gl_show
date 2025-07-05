#version 330 core
out vec4 FragColor;

in vec3 TexCoords;

uniform samplerCube SkyBox;

const float PI=3.141592;
void main()
{
    vec3 N=normalize(TexCoords);//Z
    vec3 Y=vec3(0.0,1.0,0.0); // Y 先随机一个用于计算 X
    vec3 X=normalize(cross(Y,N)); // X
    Y=normalize(cross(N,X)); // 纠正 Y

    vec3 res =vec3(0.0);
    //2 PI 1000
    float step1=2.0*PI/80.0;
    float step2=0.5*PI/20.0;
    //begin
    for(float phi=0.0;phi<2.0*PI;phi+=step1){
        for(float theta=0.0;theta<0.5*PI;theta+=step2){
            // 进行循环采样，以 法线方向为 Y 轴进行偏移
            vec3 localL=vec3(sin(theta)*cos(phi),sin(theta)*sin(phi),cos(theta));
            vec3 L=localL.x*X+localL.y*Y+localL.z*N;
            res +=texture(SkyBox, L).rgb*cos(theta)*sin(theta); // 也受角度影响
        }
    }
    //end
    res =PI* res /(80.0*20.0);
    FragColor =vec4(res,1.0);
}
