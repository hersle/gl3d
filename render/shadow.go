package render

/*

func (r *ShadowMapRenderer) SetCamera(c *camera.PerspectiveCamera) {
	r.sp.Far.Set(c.Far)
	r.sp.LightPosition.Set(c.Position)
	r.sp.ViewMatrix.Set(c.ViewMatrix())
	r.sp.ProjectionMatrix.Set(c.ProjectionMatrix())
}

func (r *ShadowMapRenderer) SetMesh(m *object.Mesh) {
	r.sp.ModelMatrix.Set(m.WorldMatrix())
}

func (r *ShadowMapRenderer) SetSubMesh(sm *object.SubMesh) {
	var v object.Vertex
	r.sp.Position.SetSource(sm.Geo.VertexBuffer(), v.PositionOffset(), v.Size())
	r.sp.SetAttribIndexBuffer(sm.Geo.IndexBuffer())
}

func (r *ShadowMapRenderer) SetCamera2(c *camera.OrthoCamera) {
	r.sp2.ViewMatrix.Set(c.ViewMatrix())
	r.sp2.ProjectionMatrix.Set(c.ProjectionMatrix())
}

func (r *ShadowMapRenderer) SetMesh2(m *object.Mesh) {
	r.sp2.ModelMatrix.Set(m.WorldMatrix())
}

func (r *ShadowMapRenderer) SetSubMesh2(sm *object.SubMesh) {
	var v object.Vertex
	r.sp2.Position.SetSource(sm.Geo.VertexBuffer(), v.PositionOffset(), v.Size())
	r.sp2.SetAttribIndexBuffer(sm.Geo.IndexBuffer())
}

// render shadow map to l's shadow map
func (r *ShadowMapRenderer) RenderPointLightShadowMap(s *scene.Scene, l *light.PointLight) {
	// TODO: re-render also when objects have moved
	if !l.DirtyShadowMap {
		return
	}

	forwards := []math.Vec3{
		math.NewVec3(+1, 0, 0),
		math.NewVec3(-1, 0, 0),
		math.NewVec3(0, +1, 0),
		math.NewVec3(0, -1, 0),
		math.NewVec3(0, 0, +1),
		math.NewVec3(0, 0, -1),
	}
	ups := []math.Vec3{
		math.NewVec3(0, -1, 0),
		math.NewVec3(0, -1, 0),
		math.NewVec3(0, 0, +1),
		math.NewVec3(0, 0, -1),
		math.NewVec3(0, -1, 0),
		math.NewVec3(0, -1, 0),
	}

	c := camera.NewPerspectiveCamera(90, 1, 0.1, l.ShadowFar)
	c.Place(l.Position)

	r.renderState.Program = r.sp.ShaderProgram

	// UNCOMMENT THIS LINE AND ANOTHER ONE TO DRAW SHADOW CUBE MAP AS SKYBOX
	//shadowCubeMap = l.shadowMap

	for face := 0; face < 6; face++ {
		r.framebuffer.AttachCubeMapFace(graphics.DepthAttachment, l.ShadowMap.Face(graphics.CubeMapLayer(face)), 0)
		r.framebuffer.ClearDepth(1)
		c.SetForwardUp(forwards[face], ups[face])

		r.SetCamera(c)

		for _, m := range s.Meshes {
			r.SetMesh(m)
			for _, subMesh := range m.SubMeshes {
				if !c.Cull(subMesh) {
					r.SetSubMesh(subMesh)

					graphics.NewRenderCommand(graphics.Triangle, subMesh.Geo.Inds, 0, r.renderState).Execute()
				}
			}
		}
	}

	l.DirtyShadowMap = false
}

func (r *ShadowMapRenderer) RenderSpotLightShadowMap(s *scene.Scene, l *light.SpotLight) {
	// TODO: re-render also when objects have moved
	if !l.DirtyShadowMap {
		return
	}

	r.framebuffer.AttachTexture2D(graphics.DepthAttachment, l.ShadowMap, 0)
	r.framebuffer.ClearDepth(1)
	r.renderState.Program = r.sp.ShaderProgram
	r.SetCamera(&l.PerspectiveCamera)

	for _, m := range s.Meshes {
		r.SetMesh(m)
		for _, subMesh := range m.SubMeshes {
			if !l.PerspectiveCamera.Cull(subMesh) {
				r.SetSubMesh(subMesh)

				graphics.NewRenderCommand(graphics.Triangle, subMesh.Geo.Inds, 0, r.renderState).Execute()
			}
		}
	}

	l.DirtyShadowMap = false
}

func (r *ShadowMapRenderer) RenderDirectionalLightShadowMap(s *scene.Scene, l *light.DirectionalLight) {
	// TODO: re-render also when objects have moved
	if !l.DirtyShadowMap {
		return
	}

	r.framebuffer.AttachTexture2D(graphics.DepthAttachment, l.ShadowMap, 0)
	r.framebuffer.ClearDepth(1)
	r.renderState.Program = r.sp2.ShaderProgram
	r.SetCamera2(&l.OrthoCamera)

	for _, m := range s.Meshes {
		r.SetMesh2(m)
		for _, subMesh := range m.SubMeshes {
			if !l.OrthoCamera.Cull(subMesh) {
				r.SetSubMesh2(subMesh)

				graphics.NewRenderCommand(graphics.Triangle, subMesh.Geo.Inds, 0, r.renderState).Execute()
			}
		}
	}

	l.DirtyShadowMap = false
}

func (r *ShadowMapRenderer) RenderShadowMaps(s *scene.Scene) {
	for _, l := range s.PointLights {
		r.RenderPointLightShadowMap(s, l)
	}
	for _, l := range s.SpotLights {
		r.RenderSpotLightShadowMap(s, l)
	}
	for _, l := range s.DirectionalLights {
		r.RenderDirectionalLightShadowMap(s, l)
	}
}
*/
